從異世界歸來的第十五天 - Kubernetes Deployment Strategies - Canary Deployment 金絲雀部署 (四)
---

## 概述

在昨天我們單純使用了 `service` 就實現出了藍綠部署，那可能就有人會心想：那說好的用 `Ingress` 實現的方法呢？是不是我隨便講講別人就隨便信信，別急別急你們仔細看看現在不就來了嗎？沒錯！繼昨天的 `service` 操作後，今天要給各位分享的是使用我們的路由守護神 `Ingress` 實現金絲雀部署的練習。

## Nginx Ingress 金絲雀部署功能介紹

`Ingress` 基於七層的 HTTP 和 HTTPS 協議進行轉發，可以通過域名和路徑對訪問做到更細粒度的劃分。`Ingress` 作為 `Kubernetes` 集群中一種獨立的資源，需要通過創建它來製定外部訪問流量的轉發規則，並通過 Ingress Controller 將其分配到一個或多個 `Service` 中。`Ingress Controller` 在不同廠商之間有著不同的實現方式，Kubernetes官方維護的 Controller 為 `Nginx Ingress` ，其支持通過配置註解（Annotations）來實現不同場景下的發布和測試。

![https://ithelp.ithome.com.tw/upload/images/20220915/2014956257Jt3Piwcc.png](https://ithelp.ithome.com.tw/upload/images/20220915/2014956257Jt3Piwcc.png)

目前 `Nginx Ingress` 提供三種基於 Header、Cookie、權重三種外部流量切分策略，只需要簡單的在註解（Annotations）寫入其提供設定即可使用：

- **nginx.ingress.kubernetes.io/canary** ：其值為 `true` 的話，將被視為 `Canary Ingress` ，為以下配置進行流量切分，使兩個 `Ingress` 互相配合。
- **nginx.ingress.kubernetes.io/canary-by-header-value** ：通知 `Ingress` 如有與 `Header` 設定值匹配的請求 `Header` ，轉導流量到 `Canary Ingress` 。
- **nginx.ingress.kubernetes.io/canary-by-header-pattern** ：運作方式與 `canary-by-header-value` 相同，並且還支持正則表達式配對之，要注意的是當 `canary-by-header-value` 如果有被設定的話，此註解的功能將會被忽略。
- **nginx.ingress.kubernetes.io/canary-by-cookie** ：通知 `Ingress` 如有與 `Cookie` 設定值匹配的請求 `Cookie` ，轉導流量到 `Canary Ingress` 。如果將其值設定為 `always` ，將會轉導所有流量。
- **nginx.ingress.kubernetes.io/canary-weight** ：此數值預設為基於零到一百的整數，代表著聲明有多少百分比的流量將會被轉導到 `Canary Ingress` 。

以上設定的優先級由高到低分別為：`canary-by-header -> canary-by-cookie -> canary-weight` 。

## 金絲雀部署策略(Canary Deployment)

金絲雀部署與藍綠部署最大的不同是，它不是非黑即白的部署方式，而是介在於黑與白之間，能夠平滑過渡到下一個版本的方法。它能夠緩辦的將修改推送給小部分的使用者，確定沒問題後才正式迭代到下一個版本，以降低值接引入新功能的風險。

以下為金絲雀部署新舊版本更新過程中接收流量的示意圖：

![https://ithelp.ithome.com.tw/upload/images/20220915/20149562NprNgfdYUP.png](https://ithelp.ithome.com.tw/upload/images/20220915/20149562NprNgfdYUP.png)

## 使用藍綠部署策略更新服務

今天我們將會使用 `Nginx Ingress` 提供的 `canary` 功能來實現，藉由簡單的在註釋中指定需要被分流的權重比例。

大致實現方法可以簡單分為以下步驟：

1. 啟動一個原有的 `v1` 版本服務並且使其與 `Ingress` 綁定成為唯一對外的正式版本。
2. 啟動並且等待我們的 `v2` 版本完全就緒，此時新舊兩個版本處於同時存在的狀態。
3. 加入 `Canary Ingress` 並設定預期分流到 `v2` 版本的請求權重。
4. 直到確認 `v2` 版本有足夠條件取代 `v1` 版本後，將 `Ingress` 指向 `v2` 版本並且關閉 `Canary Ingress`。
5. 確保終止舊的 `v1` 版本。

大致了解後就馬上來實現吧！

### 實戰練習

```yaml
# app-v1.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-deployment
  labels:
    app: my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
      version: v1
  template:
    metadata:
      labels:
        app: my-app
        version: v1
    spec:
      containers:
        - name: foo
          image: mikehsu0618/foo
          ports:
            - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: foo-service
spec:
  selector:
    app: my-app
    version: v1
  type: NodePort
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
```

```yaml
# app-v2.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar-deployment
  labels:
    app: my-app
spec:
  replicas: 1
  selector:
    matchLabels:
      app: my-app
      version: v2
  template:
    metadata:
      labels:
        app: my-app
        version: v2
    spec:
      containers:
        - name: bar
          image: mikehsu0618/bar
          ports:
            - containerPort: 8080
---
apiVersion: v1
kind: Service
metadata:
  name: bar-service
spec:
  selector:
    app: my-app
    version: v2
  type: NodePort
  ports:
    - protocol: TCP
      port: 8080
      targetPort: 8080
```

```yaml
# ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
spec:
  ingressClassName: nginx
  defaultBackend:
    service:
      name: foo-service
      port:
        number: 8080
```

1. 首先我們將運行起 `v1` 版本並且啟用 `Ingress` 服務當作我們的 `LoadBalancer` ：

```yaml
kubectl apply -f app-v1.yaml,ingress.yaml
--------
deployment.apps/foo-deployment created
service/foo-service created
ingress.networking.k8s.io/my-ingress created
```

查看 `v1` 版本服務以及 `Ingress` 狀態：

```yaml
kubectl get ingress
--------
NAME         CLASS   HOSTS   ADDRESS     PORTS   AGE
my-ingress   nginx   *       localhost   80      2m14s

============================================================================
kubectl get all
--------
NAME                                  READY   STATUS    RESTARTS   AGE
pod/foo-deployment-68df868866-hjsdx   1/1     Running   0          82s

NAME                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
service/foo-service   NodePort    10.109.223.51   <none>        8080:30256/TCP   82s
service/kubernetes    ClusterIP   10.96.0.1       <none>        443/TCP          32d

NAME                             READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/foo-deployment   1/1     1            1           82s

NAME                                        DESIRED   CURRENT   READY   AGE
replicaset.apps/foo-deployment-68df868866   1         1         1       82s
```

現在我們可以發送一些請求確認一下是否 `v1` 版本為唯一對外的服務：

```yaml
for i in {1..10}; do curl localhost; echo; done
-------
{"data":"Hello foo"}      
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"} 
```

1. 接下來就把 `v2` 版本也完整啟動：

```yaml
kubectl apply -f app-v2.yaml
----------
deployment.apps/bar-deployment created
service/bar-service created
```

查看 `v2` 版本服務是否啟動完畢：

```yaml
kubectl get all
----------
NAME                                  READY   STATUS    RESTARTS   AGE
pod/bar-deployment-7bbbff5c97-n7zhj   1/1     Running   0          52s
pod/foo-deployment-68df868866-hjsdx   1/1     Running   0          14m

NAME                  TYPE        CLUSTER-IP      EXTERNAL-IP   PORT(S)          AGE
service/bar-service   NodePort    10.97.149.224   <none>        8080:32756/TCP   52s
service/foo-service   NodePort    10.109.223.51   <none>        8080:30256/TCP   14m
service/kubernetes    ClusterIP   10.96.0.1       <none>        443/TCP          32d

NAME                             READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/bar-deployment   1/1     1            1           52s
deployment.apps/foo-deployment   1/1     1            1           14m

NAME                                        DESIRED   CURRENT   READY   AGE
replicaset.apps/bar-deployment-7bbbff5c97   1         1         1       52s
replicaset.apps/foo-deployment-68df868866   1         1         1       14m
```

確認目前依然是只開放 v1 版本接收請求：

```yaml
for i in {1..10}; do curl localhost; echo; done
-------
{"data":"Hello foo"}      
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"} 
```

成功運行～接下來將要迎接主角 `Canary Ingress` 登場。

1. 加入 `Canary Ingress` 實現請求依權重分流到新舊版本：

```yaml
# canary-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/canary: "true"
    nginx.ingress.kubernetes.io/canary-weight: "10"
  name: canary-ingress
spec:
  ingressClassName: nginx
  defaultBackend:
    service:
      name: bar-service
      port:
        number: 8080
```

此處我們設定將 10% 比重的請求分流到 `bar-service` 這個 `v2` 版本的服務。

建立 `Canary Ingress` 資源：

```yaml
kubectl apply -f canary-ingress.yaml
---------
ingress.networking.k8s.io/canary-ingress created
```

此時我們可以預期對 `localhost` 的請求中有百分之十會由 `v2` 版本接收：

```yaml
for i in {1..10}; do curl localhost; echo; done
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello bar"} // 出現ㄌ！！
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
{"data":"Hello foo"}
```

太神啦～冰鳥，到此我們已經輕鬆的實現金絲雀部署了。

接下來我們只剩下把等待條件許可時，將 `Ingress` 設定為新版本的 `v2` 並且刪除過渡使用的 `Canary Ingress` 以及 `v1` 版本資源即可。

## 結論

我們終於完成了進階部署策略篇，也藉由了部署這個大觀念不斷重複加深了 `Deployment` `Service` `Pod` `Ingress` 這幫好兄弟的使用方法，再次恭喜堅持走到這個同學還有我自己（嗚嗚，這一切都得來不易。老話一句關於部署的方式完全沒有侷限於任何形式，尤其在工作上我們有更多需要顧慮的 X 因子，各種花式奇淫技巧推層出新為了都是想應付某種特定業務情境，我們唯一能做到的就是穩紮穩打以不變應萬變，把基礎概念打好才是解決問題的根本。話說可以實現第七層負載均衡的 `Ingress Controller` ，事實上一門非常大的學問，而對 `LoadBalancer` 更深入了解一定也是學習 `Kubernetes` 的重要課題，希望日後也能有機會做一個更深入的探討。


相關文章：

- [從異世界歸來的第三天 - Kubernetes 的組件](https://ithelp.ithome.com.tw/articles/10287576)

- [從異世界歸來的第六天 - Kubernetes 三兄弟 - 實戰做一個 Pod (一)](https://ithelp.ithome.com.tw/articles/10288199)

- [從異世界歸來的第七天 - Kubernetes 三兄弟 - 實戰做一個 Service (二)](https://ithelp.ithome.com.tw/articles/10288389)

- [從異世界歸來的第八天 - Kubernetes 三兄弟 - 實戰做一個 Deployment (三)](https://ithelp.ithome.com.tw/articles/10288602)

- [從異世界歸來的第十二天 - Kubernetes Deployment Strategies - 常見的部署策略 (一)](https://ithelp.ithome.com.tw/articles/10289496)


相關程式碼同時收錄在：

[https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day15](https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day15)

Reference

****[什麼是藍綠部署、金絲雀部署、滾動部署、紅黑部署、AB測試？](https://www.gushiciku.cn/pl/gUOs/zh-tw)****

****[管理資源](https://kubernetes.io/zh-cn/docs/concepts/cluster-administration/manage-deployment/#canary-deployments)****

[ContainerSolutions/**k8s-deployment-strategies**](https://github.com/ContainerSolutions/k8s-deployment-strategies/tree/master/canary)

**[K8S学习笔记之Kubernetes 部署策略详解](https://cloud.tencent.com/developer/article/1411271)**

****[使用Nginx Ingress實現灰度發布和藍綠髮布](http://dockone.io/article/2434773)****

**[Configure a canary deployment](https://docs.mirantis.com/mke/3.5/ops/deploy-apps-k8s/nginx-ingress/configure-canary-deployment.html)**

[NGINX Ingress Controller Annotations](nginx.ingress.kubernetes.io/canary-by-cookie)