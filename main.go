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
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"context"
)

func main() {
	CONFIG_MAP_NAME :="app-config"
	log.Println("Starting Server ....")
	r := mux.NewRouter()
	r.HandleFunc("/", index)
	r.HandleFunc("/hello", hello).Methods("GET")
	r.HandleFunc("/config", config_handeler).Methods("GET")
	r.HandleFunc("/data", data_handeler).Methods("GET")
	http.Handle("/", r)

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

	// loop killer
	go watchrepo(CONFIG_MAP_NAME)

	log.Println("Web Server is Starting on 6060 Port ....")
	log.Fatal(http.ListenAndServe(":6060", nil))
}


func config_handeler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received a request for config_handeler")
	
	query := r.URL.Query()

	github_username := query.Get("github_username")
	if github_username == "" {
		github_username = "default"
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
	log.Printf("Start Loop killer \n")
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
		go call_api(CONFIG_MAP_NAME)
		log.Printf("Sleeping for %s seconds \n", interval)
		time.Sleep(time.Duration(wait) * time.Second)
	}
}

// A Repo Struct to map every Repo to.
type Repo struct {
    Name string	`json:"name"`
    URL  string	`json:"html_url"`
}

func call_api(CONFIG_MAP_NAME string) {
	
	api_url := "https://api.github.com/users/yassermog/repos"
	
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
	json.Unmarshal([]byte(responseData), &repos)

    for i := 0; i < len(repos); i++ {
        log.Printf("Repo name %s\n",repos[i].Name)
        log.Printf("Repo url %s\n",repos[i].URL)
        log.Printf("###################################################")
    }
}

func data_handeler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received a request for config_handeler")
	// to change
	configMapName := "app-config"
	namespace := "default"
	getData(namespace,configMapName)
}

func getData(namespace string,configMapName string){
	
	client, err := newClient("")
	if err != nil {
		log.Fatal(err)
	}
	configMapData := make(map[string]string, 0)		
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

	var cm *corev1.ConfigMap
	if _, err := client.CoreV1().ConfigMaps("default").Get(context.Background(),"app-config", metav1.GetOptions{}); errors.IsNotFound(err) {
		//cm, _ = client.CoreV1().ConfigMaps("default").Create(context.Background(),&configMap,metav1.CreateOptions{})
		log.Printf( "No Data %v \n",cm)
	} else {
		cm, _ = client.CoreV1().ConfigMaps("default").Update(context.Background(),&configMap,metav1.UpdateOptions{})
		log.Printf( "there is data %v \n",cm)
	}
	log.Printf( " config map %+v  \n",cm)


	repos :=os.Getenv("repos")
	log.Printf( "value env %v  \n",repos)

}


func setData(namespace string,configMapName string){
	client, err := newClient("")
	if err != nil {
		log.Fatal(err)
	}
		
	configMapData := make(map[string]string, 0)
	uiProperties := `
					color.good=purple
					color.bad=yellow
					allow.textmode=true
					`
	fmt.Printf("configMapData %v", configMapData)
				
	configMapData["ui.properties"] = uiProperties
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

	var cm *corev1.ConfigMap
	if _, err := client.CoreV1().ConfigMaps("default").Get(context.Background(),"app-config", metav1.GetOptions{}); errors.IsNotFound(err) {
	cm, _ = client.CoreV1().ConfigMaps("default").Create(context.Background(),&configMap,metav1.CreateOptions{})
	} else {
	cm, _ = client.CoreV1().ConfigMaps("default").Update(context.Background(),&configMap,metav1.UpdateOptions{})
	}
	log.Printf( " config map %+v  \n",cm)


os.Getenv("BAR")
	log.Printf( "value env %v  \n",cm)


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

func hello(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	name := query.Get("name")
	if name == "" {
		name = "Guest"
	}
	log.Printf("Received request for %s\n", name)
	w.Write([]byte(fmt.Sprintf("Hello, %s\n", name)))
}

func index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Home Page ! \n")
	fmt.Fprintf(w, "Try /config  : to config the repo watcher \n")
}