# repo_watcher
repo_watcher

## Deployment

To deploy to kubernetes : 


Checkout the repository and run 

```
helm install repo-watcher ./repo-watcher/ --set service.type=NodePort
```
or 

```
make install
```


to Authrize the pod access the default namespace 
```
make auth
```

## Screens 

Pod logs 

![Alt text](img/firstime.PNG)

no updates 

![Alt text](img/noupdate.PNG)

new repo 

![Alt text](img/newrepo.PNG)


changing github username 

![Alt text](img/newusernme.png)

## configMap 

to save the data of the reositories I used a configmap in the helm chart  

```
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: default
data:
  repos: ""
```