package main

import (
    "fmt"
    "log"
	"os"
    "net/http"
	"github.com/gorilla/mux"
	"gopkg.in/natefinch/lumberjack.v2"
    "time"
	"io/ioutil"
	"strconv"
	"encoding/json"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"context"
	"strings"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type RepoCollector struct {
	RepoMetric *prometheus.Desc
}

var  numberOfRepos float64

func newRepoCollector() *RepoCollector {
	return &RepoCollector{
		RepoMetric: prometheus.NewDesc("available_user_repositories",
			"Shows the number available user repositories",
			nil, nil,
		),
	}
}
func (collector *RepoCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.RepoMetric
}
func (collector *RepoCollector) Collect(ch chan<- prometheus.Metric) {
	var metricValue float64
	metricValue=numberOfRepos
	
	ch <- prometheus.MustNewConstMetric(collector.RepoMetric, prometheus.CounterValue, metricValue)
}
func main() {
	CONFIG_MAP_NAME :="app-config"
	log.Println("Starting Server ....")
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/config", config_handeler).Methods("GET")
	r.HandleFunc("/data", data_handeler).Methods("GET")
	http.Handle("/", r)
	http.Handle("/metrics", promhttp.Handler())


	// Configure Logging
	LOG_FILE_LOCATION := os.Getenv("LOG_FILE_LOCATION")
	if LOG_FILE_LOCATION != "" {
		log.SetOutput(&lumberjack.Logger{
			Filename:   LOG_FILE_LOCATION,
			MaxSize:    500, // megabytes
			MaxBackups: 3,
			MaxAge:     28,   //days
			Compress:   true, // disabled by default
		})
	}

	// configration 
	interval := "40"
	
	os.Setenv("watch_interval", interval)

	// loop watchrepo
	go watchrepo(CONFIG_MAP_NAME)

	numberOfReposMetric := newRepoCollector()
  	prometheus.MustRegister(numberOfReposMetric)

	log.Println("Web Server is Starting on 6060 Port ....")
	log.Fatal(http.ListenAndServe(":6060", nil))
}


func config_handeler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received a request for config_handeler")
	
	query := r.URL.Query()

	github_username := query.Get("github_username")
	if github_username == "" {
		github_username = "yassermog"
	} else{
		fmt.Fprintf(w, "github_username is set to = %s\n", github_username)
		setConfigMap("")
	}
	os.Setenv("github_username", github_username)

	interval := query.Get("interval")
	if interval == "" {
		interval = "60"
	}
	os.Setenv("watch_interval", interval)
	
	fmt.Fprintf(w, "interval = %s\n", interval)
	fmt.Fprintf(w, "github_username = %s\n", github_username)
	
	log.Printf("interval = %s\n", interval)	
	log.Printf("github_username = %s\n", github_username)	
}

func watchrepo(CONFIG_MAP_NAME string){
	log.Printf("####################### Repository Watcher  ####################### \n")
	interval := ""
	wait := 0

	for{
		interval = os.Getenv("watch_interval")
		i, err := strconv.Atoi(interval)
		wait=i;
		if err != nil {
			log.Fatal(err)
		}
		go call_api()

		time.Sleep(time.Duration(wait) * time.Second)
	}
}

// A Repo Struct to map every Repo to.
type Repo struct {
    Name string	`json:"name"`
    URL  string	`json:"html_url"`
}

func call_api() {
	log.Printf("calling the api ..... \n")
	
	github_username := os.Getenv("github_username")

	if github_username == "" {
		github_username = "yassermog"
	}
	
	api_url := "https://api.github.com/users/"+github_username+"/repos"
	
	response, err := http.Get(api_url)
    if err != nil {
        fmt.Print(err.Error())
        os.Exit(1)
    }

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        log.Fatal(err)
    }

    var repos []Repo
    ResponseNamesArr := []string{}
	json.Unmarshal([]byte(responseData), &repos)
	numberOfRepos=float64(len(repos))
	log.Printf("getting  %d repos \n",len(repos))
	if(len(repos)==0){
		log.Printf("###################################################")
		log.Printf(" No Repos or You Reach the limit of github")
		log.Printf("###################################################")
	}else{
		for i := 0; i < len(repos); i++ {
			n := repos[i].Name
			log.Printf("Repo name %s\n",n)
			ResponseNamesArr = append(ResponseNamesArr,n)
		}
		log.Printf("###################################################")
		/// compare with the old data
		ResponseNames := strings.Join(ResponseNamesArr,";")
		previousRepos := getPreviousData()
		
		
		if(previousRepos==""){
			log.Printf("################## First name running ###############")
			log.Printf("setting config map :")
			setConfigMap(ResponseNames)
		
		} else { //compare 
			if(previousRepos==ResponseNames){
				log.Printf("################## there is no updates ###############")
			} else {
				new_arr:=strings.Split(ResponseNames, ";")
				old_arr:=strings.Split(previousRepos, ";")
				diffRepos := difference(new_arr,old_arr);

				log.Printf("#################### new repos ####################")
				for i := 0; i < len(diffRepos); i++ {
					n := diffRepos[i]
					log.Printf("New Repo name %s\n",n)
				}
				log.Printf("###################################################")
				setConfigMap(ResponseNames)
			}
		}
	} 
}

// difference beteween 2 arrays
func difference(a, b []string) []string {
    mb := make(map[string]struct{}, len(b))
    for _, x := range b {
        mb[x] = struct{}{}
    }
    var diff []string
    for _, x := range a {
        if _, found := mb[x]; !found {
            diff = append(diff, x)
        }
    }
    return diff
}

func data_handeler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received a request for data_handeler")
	data :=getPreviousData()
	new_arr:=strings.Split(data, ";")
	fmt.Fprintf(w,"######################### Repos ##########################\n")

	for i := 0; i < len(new_arr); i++ {
		n := new_arr[i]
        fmt.Fprintf(w,"Repo name %s\n",n)
    }
	fmt.Fprintf(w,"###################################################\n")

}

func getPreviousData() string {
	client, err := newClient("")
	
	if err != nil {
		log.Fatal(err)
	}
	
	var cm *corev1.ConfigMap
	cm, err = client.CoreV1().ConfigMaps("default").Get(context.Background(),"app-config", metav1.GetOptions{})
	return cm.Data["repos"];
}


func setConfigMap(reposData string){
	client, err := newClient("")
	
	if err != nil {
		log.Fatal(err)
	}
	
	configMapData := make(map[string]string, 0)
	configMapData["repos"] = reposData			

	configMap := corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "app-config",
			Namespace: "default",
		},
		Data: configMapData,
	}
	
	client.CoreV1().ConfigMaps("default").Update(context.Background(),&configMap,metav1.UpdateOptions{})
	
}

func newClient(contextName string) (kubernetes.Interface, error) {
	configOverrides := &clientcmd.ConfigOverrides{CurrentContext: contextName}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	config, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides).ClientConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(config)
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page ! \n")
	fmt.Fprintf(w, "Try /config  : to config the repo watcher \n")
}