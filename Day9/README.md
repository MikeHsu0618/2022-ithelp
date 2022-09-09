從異世界歸來的第九天 - Kubernetes 老大哥 - 路由守護神 Ingress
----

## 概述

在前兩天的 `Kubernetes 三兄弟 - 實戰做一個 Service (二)` 我們介紹了 `Service` 這個元件，並且在利用他讓集群中的 `Pod` 可以被外部的我們存取。然而每個物件都需要指定對外的 `port number` 以及 Node 上的 `port mapping` ，這就代表 `愈多的 Service 就要管理愈多的 port number` ，而且現在任何的網站上面如果還需要加上 `port number` 實用性實在大打折扣。

## 什麼是 Ingress？

Ingress 可以幫我們統一對外的 `port number` ，並且根據 hostname 或是 pathname 決定請求要轉發到哪個 `Service` 上成為更上層的 `LoadBalancer` ，並且 `Kubernetes Ingress` 會統一打開 http 的 80 port 以及 https 的 443 port，解決剛才提到的 port number 紊亂不一的問題。


![https://ithelp.ithome.com.tw/upload/images/20220909/201495626AeXVmPUtc.png](https://ithelp.ithome.com.tw/upload/images/20220909/201495626AeXVmPUtc.png)

## Ingress 作用

Ingress 負責的事情主要被定義為下面幾項：

- 將不同路徑的請求對應到各自的 Service（give services externally-reachable urls ）：只要透過設定好的 hostname 跟 pathname 就可以觸及到對應的 Services 進而存取其對應的 Pods。
- 流量的負載均衡（load balance traffic）：例如負載均衡算法、後端權重方案等。
- 支持 SSL Termination：支援 https 的傳輸層安全協定並且擔任起解密的責任，使 Service 與 Pod 之間的溝通都是以無加密方式傳輸，得以正常傳輸資料。
- 支持虛擬網域設定（offer name based virtual hosting）：Ingress 提供我們在同個 IP 下設定自己的虛擬網域，也就是我們前面提到的 `hostname` 。

## 安裝 Ingress

在本地的 `docker-desktop` 上我們只需要運行下列設定檔，Kubernetes 就會幫我們建立一個 `ingress-nginx` 的 namespace，並且運行起相關服務：

```jsx
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.2.1/deploy/static/provider/cloud/deploy.yaml

------------------------------
namespace/ingress-nginx unchanged
serviceaccount/ingress-nginx configured
serviceaccount/ingress-nginx-admission configured
role.rbac.authorization.k8s.io/ingress-nginx configured
role.rbac.authorization.k8s.io/ingress-nginx-admission configured
clusterrole.rbac.authorization.k8s.io/ingress-nginx configured
clusterrole.rbac.authorization.k8s.io/ingress-nginx-admission configured
rolebinding.rbac.authorization.k8s.io/ingress-nginx configured
rolebinding.rbac.authorization.k8s.io/ingress-nginx-admission configured
clusterrolebinding.rbac.authorization.k8s.io/ingress-nginx configured
clusterrolebinding.rbac.authorization.k8s.io/ingress-nginx-admission configured
configmap/ingress-nginx-controller configured
service/ingress-nginx-controller created
service/ingress-nginx-controller-admission created
deployment.apps/ingress-nginx-controller created
job.batch/ingress-nginx-admission-create created
job.batch/ingress-nginx-admission-patch created
ingressclass.networking.k8s.io/nginx configured
validatingwebhookconfiguration.admissionregistration.k8s.io/ingress-nginx-admission configured
```

檢查是否成功運作：

```jsx
kubectl get all -n ingress-nginx

------------------------------
NAME                                            READY   STATUS      RESTARTS   AGE                                                      
pod/ingress-nginx-admission-create-rf8cl        0/1     Completed   0          7m4s
pod/ingress-nginx-admission-patch-mzmc8         0/1     Completed   0          7m4s
pod/ingress-nginx-controller-778667bc4b-twt6n   1/1     Running     0          7m4s

NAME                                         TYPE           CLUSTER-IP       EXTERNAL-IP   PORT(S)                      AGE
service/ingress-nginx-controller             LoadBalancer   10.105.184.158   localhost     80:30205/TCP,443:31820/TCP   7m5s
service/ingress-nginx-controller-admission   ClusterIP      10.106.69.252    <none>        443/TCP                      7m4s

NAME                                       READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/ingress-nginx-controller   1/1     1            1           7m4s

NAME                                                  DESIRED   CURRENT   READY   AGE
replicaset.apps/ingress-nginx-controller-778667bc4b   1         1         1       7m4s

NAME                                       COMPLETIONS   DURATION   AGE
job.batch/ingress-nginx-admission-create   1/1           5s         7m4s
job.batch/ingress-nginx-admission-patch    1/1           5s         7m4s
```

## 實際練習

關於 `Ingress` 的實際應用，官方有提供幾種方式讓我們用 `URL` 控制並連接到我們指定的服務。

### 1. 單一 Service

現有的 Kubernetes 允許我們直接暴露 `單一個 Service` 。現在我們依然可以透過 ingress 的 `defaultBackend` 辦到這件事，代表規範條件以外的流量通通都會遵守 `defaultBackend` 規則分配到對應服務。

準備相關設定檔：

```jsx
// deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-deployment
  labels:
    type: demo
spec:
  replicas: 2
  selector:
    matchLabels:
      type: demo
  template:
    metadata:
      labels:
        type: demo
    spec:
      containers:
        - name: foo
          image: mikehsu0618/foo
          ports:
            - containerPort: 8080
```

```jsx
// service.yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  selector:
    type: demo
  type: NodePort // 這裡我們將能直接暴露端口的 `Loadbalancer` 改成 `NodePort`
  ports:
    - protocol: TCP
      port: 8000
      targetPort: 8080
      nodePort: 30390
```

```jsx
// ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
spec:
  ingressClassName: nginx
  defaultBackend:
    service:
      name: my-service
      port:
        number: 8000
```

在上面的 `ingress.yaml` 我們設定了 `defaultBackend` 讓所有流量都預設導到 `my-service` 中的 `8000 port`，形成了一條從 `loabalancer` → `services` → `pods` 的路徑。

執行以上設定檔：

```jsx
kubectl apply -f deployment.yaml,service.yaml,ingress.yaml
----------------------------

deployment.apps/foo-deployment unchanged
service/my-service unchanged
ingress.networking.k8s.io/my-ingress unchanged
```

查看服務狀況：

```jsx
kubectl get all
----------------------------

NAME                                  READY   STATUS    RESTARTS   AGE
pod/foo-deployment-6bbf665b47-6769n   1/1     Running   0          41m
pod/foo-deployment-6bbf665b47-96khw   1/1     Running   0          41m

NAME                 TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)          AGE
service/kubernetes   ClusterIP   10.96.0.1      <none>        443/TCP          27d
service/my-service   NodePort    10.108.203.7   <none>        8000:30390/TCP   41m

NAME                             READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/foo-deployment   2/2     2            2           41m

NAME                                        DESIRED   CURRENT   READY   AGE
replicaset.apps/foo-deployment-6bbf665b47   2         2         2       41m
```

查看 Ingress ：

```jsx
kubectl get ingress
----------------------------

NAME         CLASS   HOSTS   ADDRESS     PORTS   AGE
my-ingress   nginx   *       localhost   80      42m
```

成功將啟動了一個 `[localhost:80](http://localhost:80)` 的負載均衡器～

實際測試：

```jsx
curl localhost
----------------------------

{"data":"Hello foo"}
```

### 2. Simple Fanout and Visual hosting

一個 `fanout` 可以根據請求的 URL 將來自同一個 IP 地址的流量轉到到多個 Service。並且實現以下配置：

![https://ithelp.ithome.com.tw/upload/images/20220909/20149562cKcYfu1xsc.png](https://ithelp.ithome.com.tw/upload/images/20220909/20149562cKcYfu1xsc.png)

準備相關設定檔：

```jsx
// deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-deployment
  labels:
    type: foo-demo
spec:
  replicas: 2
  selector:
    matchLabels:
      type: foo-demo
  template:
    metadata:
      labels:
        type: foo-demo
    spec:
      containers:
        - name: foo
          image: mikehsu0618/foo
          ports:
            - containerPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar-deployment
  labels:
    type: bar-demo
spec:
  replicas: 1
  selector:
    matchLabels:
      type: bar-demo
  template:
    metadata:
      labels:
        type: bar-demo
    spec:
      containers:
        - name: bar
          image: mikehsu0618/bar
          ports:
            - containerPort: 8080
```

```jsx
// service.yaml
apiVersion: v1
kind: Service
metadata:
  name: foo-service
spec:
  type: NodePort
  selector:
    type: foo-demo
  ports:
    - protocol: TCP
      port: 8000
      targetPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: bar-service
spec:
  type: NodePort
  selector:
    type: bar-demo
  ports:
    - protocol: TCP
      port: 8000
      targetPort: 8080
```

```jsx
// ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
spec:
  ingressClassName: nginx
  rules:
    - host: foo.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: foo-service
                port:
                  number: 8000
    - host: bar.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: bar-service
                port:
                  number: 8000
```

可以在上面的設定檔看到我們用 `ingress` 產生出的 `virsual hosting` 並且使不同的網域對應到不同的 `Service` ，實現預期的 `fanout` 效果。

讓我們運行以上設定檔：

```jsx
kubectl apply -f deployment.yaml,service.yaml,ingress.yaml

deployment.apps/foo-deployment configured
deployment.apps/bar-deployment configured
service/foo-service configured
service/bar-service configured
ingress.networking.k8s.io/my-ingress configured
```

查看 ingress 詳細資訊：

```jsx
kubectl describe ingress my-ingress
----------------------------

Name:             my-ingress                     
Labels:           <none>
Namespace:        default
Address:          localhost
Ingress Class:    nginx
Default backend:  <default>
Rules:
  Host        Path  Backends
  ----        ----  --------
  foo.com     
              /   foo-service:8000 (10.1.1.169:8080,10.1.1.170:8080)
  bar.com     
              /   bar-service:8000 (10.1.1.168:8080)
Annotations:  <none>
Events:
  Type    Reason  Age                  From                      Message
  ----    ------  ----                 ----                      -------
  Normal  Sync    7m35s (x4 over 15m)  nginx-ingress-controller  Scheduled for sync
```

Ingress 已經成功的替我們架起了 `[foo.com](http://foo.com)` `[bar.com](http://bar.com)` 兩個虛擬網域，並幫我們把服務連接到對應的 `Service` 上～

因為我們是在本地上架起虛擬網域的，所以我們需要讓以上兩個網域反向代理到本地 `127.0.0.1` 上，所以我們這時必須去調整 `/etc/hosts` 檔本地才能順利吃到路徑請求：

```jsx
sudo vim /etc/hosts

// 在檔案中插入以下需要映射的網域
127.0.0.1 foo.com
127.0.0.1 bar.com

// 在鍵盤中手動輸入下列字元來儲存！
:wq!
```

接著來實際測試：

```jsx
curl http://foo.com
{"data":"Hello foo"}

curl http://bar.com
{"data":"Hello bar"}
```

大功告成！

## 結論

感謝願意看到這裡的同鞋們，到這裡我們可以說是已經初窺 `Kubernetes` 的門徑，熟悉 docker 的人已經有能力可以在本地 run 起自己想要的服務，並且配置 `URL` 路徑實現負載均衡。說說我自己的收穫，因為接觸了 `Kubernetes` 讓我開始學習思考如何實現一套微服務系統，他的出現對我這個之前總是在寫單體式應用的小廢廢來說，對分佈式架構有個更清晰的輪廓並且深深著迷，還有太多東西可以學習了，就讓我們繼續堅持下去吧。

相關文章：
- [從異世界歸來的第三天 - Kubernetes 的組件](https://ithelp.ithome.com.tw/articles/10287576)

- [從異世界歸來的第六天 - Kubernetes 三兄弟 - 實戰做一個 Pod (一)](https://ithelp.ithome.com.tw/articles/10288199)

- [從異世界歸來的第七天 - Kubernetes 三兄弟 - 實戰做一個 Service (二)](https://ithelp.ithome.com.tw/articles/10288389)

- [從異世界歸來的第八天 - Kubernetes 三兄弟 - 實戰做一個 Deployment (三)](https://ithelp.ithome.com.tw/articles/10288602)


相關程式碼同時收錄在：
https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day9

Reference

[Kubernetes Documentation-ingress](https://kubernetes.io/docs/concepts/services-networking/ingress/)

**[[Kubernetes] Resource Object 概觀](https://godleon.github.io/blog/Kubernetes/k8s-CoreConcept-ResourceObject-Overview/)**

****[Kubernetes 那些事 — Ingress 篇（一）](https://medium.com/andy-blog/kubernetes-%E9%82%A3%E4%BA%9B%E4%BA%8B-ingress-%E7%AF%87-%E4%B8%80-92944d4bf97d)****