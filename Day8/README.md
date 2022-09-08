從異世界歸來的第八天 - Kubernetes 三兄弟 - 實戰做一個 Deployment (三)
---

## 概述

今天要介紹的是 Kubernetes 三兄弟的 `Deployment`，這個資源對象為 `Pod` 和 `ReplicaSet` 兩者提供了一個聲明式（declarative）定義的方法來達到使用者所期望的容器執行狀態，並且官方建議透過 Deployment 來佈署 `Pod` 和 `ReplicaSet` ，典型的應用場景包括：

- 定義 Deployment 來創建Pod 和ReplicaSet
- 滾動升級和回滾應用
- 擴容和縮容
- 暫停和繼續 `Deployment`

`Pod` 的介紹相信大家已經都不陌生了，但這邊怎麼又冒出一個 `ReplicaSet` 呢？ `ReplicaSet` 是用來確保在資源允許的前提下，指定的 pod 的數量會跟使用者期望的一致，也就是所謂的 `desired status` ，而官方建議 `ReplicaSet` 要搭配 `Deployment` 一起來使用是因為 `Deployment` 是個更上層的抽象概念，也支援了更多好用的功能，因此官方才會建議不要單獨使用 `ReplicaSet` ，而是使用 `Deployment` 並且將其相關資訊設定在裡面。

從下圖可以看出三者在 Kubernetes 中的對應關係：

![https://ithelp.ithome.com.tw/upload/images/20220908/20149562Bf86dCdeZS.png](https://ithelp.ithome.com.tw/upload/images/20220908/20149562Bf86dCdeZS.png)

## 使用案例

官方貼心的為我們提供了幾個經典的 `Deployment` 使用案例：

- 使用 Deployment 來創建 ReplicaSet，而 ReplicaSet 在後台創建 Pod 並檢查成功或失敗。
- 更新 Deployment 的 Pod 設定來聲明 Pod 的新狀態。這會創建一個新的 ReplicaSet ，Deployment 將會按照控制速率（controlled rate）將 Pod 裝態更新至新的 ReplicaSet 設定。
- 回滾到先前的 Deployment 版本，如果當前的版本不穩定。
- 擴展或收縮 Deployment 以承載更多負荷。

接下來我們將用以上情境來實戰演練一下～

## 實戰演練

### 1. 創建 Deployment

```bash
// deployment.yaml

apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-deployment
  labels:
    type: demo
spec:
  replicas: 1
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar-deployment
  labels:
    type: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      type: demo
  template:
    metadata:
      labels:
        type: demo
    spec:
      containers:
        - name: bar
          image: mikehsu0618/bar
          ports:
            - containerPort: 8080
```

- **`kind` :** kind 選擇為 `Deployment`
- **`spec.replicas` :** 被選擇套用的 Container 需要產生多少個 Pod，也是我們實現水平擴展的關鍵**。**
- **`spec.selector.matchLabels` :** 這裡就是寫入需要套用此 `Deployment` 的 `Template Labels` ，所以兩者必須相同。
- **`spec.template.metadata.labels`：**設定 `template.spec` 的 `Lables` 。
- **`spec.template.spec.containers` :**  這裡就是我們熟悉的 Pod 相關設定。

接著讓我們運行設定（設定檔沒有錯誤則可以如預期中的建立）：

```bash
kubectl apply -f ./deployment.yaml

--------------------
deployment.apps/foo-deployment created
deployment.apps/bar-deployment created
```

使用指令確認一下：

```bash
kubectl get all

--------------------
NAME                                  READY   STATUS    RESTARTS   AGE
pod/bar-deployment-75bcfbd655-g5gwm   1/1     Running   0          5m59s
pod/foo-deployment-6bbf665b47-kfvxr   1/1     Running   0          5m59s

NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   23d

NAME                             READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/bar-deployment   1/1     1            1           5m59s
deployment.apps/foo-deployment   1/1     1            1           5m59s

NAME                                        DESIRED   CURRENT   READY   AGE
replicaset.apps/bar-deployment-75bcfbd655   1         1         1       5m59s
replicaset.apps/foo-deployment-6bbf665b47   1         1         1       5m59s
```

看到我們成功的運行起了 `foo` `bar` 兩個 Pod，並且建立了各自的 `Deployment` `ReplicaSet` 。

### 2. 更新 Deployment 實現水平擴展

接下來我們使用來使用不同的方法更新已經運行起來的 `Deployment` 。

直接修改原有的設定檔：

```bash
apiVersion: apps/v1
kind: Deployment
metadata:
  name: foo-deployment
  labels:
    type: demo
spec:
	// 這裡我們將 Pod 擴展成兩個！
	===================
  replicas: 2
  ===================
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar-deployment
  labels:
    type: demo
spec:
  replicas: 1
  selector:
    matchLabels:
      type: demo
  template:
    metadata:
      labels:
        type: demo
    spec:
      containers:
        - name: bar
          image: mikehsu0618/bar
          ports:
            - containerPort: 8080
```

修改完後再執行一次 `apply` 指令，kubectl 會檢查指定設定檔是否有更新：

```bash
kubectl apply -f ./deployment.yaml --record // --record 可以紀錄 rollout 歷史變更指令
-------------------------
deployment.apps/foo-deployment configured // 有更新
deployment.apps/bar-deployment unchanged  // 未檢查到更新
```

接著可以 `kubectl rolloout status` 查看我們對 `foo-deployment` 的資源管理狀態：

```bash
kubectl rollout status deployment foo-deployment 
-------------------------
deployment "foo-deployment" successfully rolled out
```

當指令顯示成功，即代表剛剛的更新已經正式生效～，但只要遇到設定錯誤或者是無法實現的請求時， `rollout status` 將會持續等待至 timeout。

我們也可以使用第二個方法「指令更新」來調整 `Deployment` ：

```bash
kubectl scale deployment bar-deployment --replicas 3
```

而第三個方法為直接編輯在 Kubernetes 運行中的 `Deployment` 設定：

```bash
// 打開 commmand 編輯面板，直接修改設定
kubectl edit deploy bar-deployment
```

來使用 `get all` 確認看看吧。

```bash
kubectl get all
-------------------------
NAME                                  READY   STATUS    RESTARTS   AGE
pod/bar-deployment-75bcfbd655-75qcd   1/1     Running   0          31s
pod/bar-deployment-75bcfbd655-c5h9w   1/1     Running   0          31s
pod/bar-deployment-75bcfbd655-g5gwm   1/1     Running   0          5h52m
pod/foo-deployment-6bbf665b47-45c2k   1/1     Running   0          4h7m
pod/foo-deployment-6bbf665b47-kfvxr   1/1     Running   0          5h52m

NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   23d

NAME                             READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/bar-deployment   3/3     3            3           5h52m
deployment.apps/foo-deployment   2/2     2            2           5h52m

NAME                                        DESIRED   CURRENT   READY   AGE
replicaset.apps/bar-deployment-75bcfbd655   3         3         3       5h52m
replicaset.apps/foo-deployment-6bbf665b47   2         2         2       5h52m
```

我們在返回結果中可以看到 `pod/bar-deployment`  已經預期的啟動三個，並且 `RepolicaSet` 和 `Deployment` 也更新了對應狀態。

### 3. 使用 Rollout 查看歷史版本並回滾

在我們更新 `Deployment` 時，Kubernetes 會產生一個 `Deployment Revision` ****，可以很簡單的理解為是更新歷史版本，但要注意的是 `不是每一次的更新都會產生 Revision` ，只有在 `Deployment created` 以及 `spec.template` ****範圍下的設定有更新才會產生，所以我們上面更新的 `replicas=3` 並不會出現在歷史中。

讓我們改動 `spec.template` 來實驗看看：

```bash
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
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: bar-deployment
  labels:
    type: demo
spec:
  replicas: 3
  selector:
    matchLabels:
      type: demo
  // 只有在 spec.template 下的改定才會紀錄在 rollout history 中
  template:
    metadata:
      labels:
        type: demo
    spec:
      containers:
        - name: bar
			// 將我們的 image tag 版號改成不存在 `v1`
      ===================================
          image: mikehsu0618/bar:v1
      ===================================
          ports:
            - containerPort: 8080
```

更新 `Deployment` 設定檔並使用 `--record` 來紀錄指令：

```bash
kubectl apply -f deployment.yaml --record
-------------------------
Flag --record has been deprecated, --record will be removed in the future
deployment.apps/foo-deployment configured
deployment.apps/bar-deployment configured
```

原本的 `spec.template` 雖然會被紀錄在 rollout history 中，但不會有額外資訊，--record 可以讓 Kubernetes 幫我們記下我們當下改變設定的那個指令。

<aside>
💡 目前 `--record` 顯示為將背棄用的 flag，但官方並沒有推出替代方案，所以大部分網友依然繼續使用 `--record`

</aside>

這時我們就能在 `rollout history` 查看產生出來的 `revision` ：

```bash
kubectl rollout history deployment bar-deployment
--------------------------
REVISION  CHANGE-CAUSE
1         <none>
2         kubectl apply --filename=deployment.yaml --record=true
```

第一個版本為先前 `Deployment` 被建立時且沒有輸入 `--record` 的版本，第二個版本為我們調整 `bar image=mikehsu0618/bar:v1` 且有 `--record` 的版本。

指定 `revision` 並查看詳細資訊：

```bash
kubectl rollout history deployment bar-deployment --revision=2
--------------------------
deployment.apps/bar-deployment with revision #2ment --revision=2
Pod Template:
  Labels:       pod-template-hash=864b65d8b6
        type=demo
  Annotations:  kubernetes.io/change-cause: kubectl apply --filename=deployment.yaml --record=true
  Containers:
   bar:
    Image:      mikehsu0618/bar:v1
    Port:       8080/TCP
    Host Port:  0/TCP
    Environment:        <none>
    Mounts:     <none>
  Volumes:      <none>
```

接下來一樣是使用 `get all` 指令查看容器狀況：

```bash
kubectl get all
--------------------------
NAME                                  READY   STATUS             RESTARTS   AGE
pod/bar-deployment-75bcfbd655-5b9z5   1/1     Running            0          7m5s
pod/bar-deployment-75bcfbd655-6whzr   1/1     Running            0          7m5s
pod/bar-deployment-75bcfbd655-zk88d   1/1     Running            0          7m5s
pod/bar-deployment-864b65d8b6-wdhz9   0/1     ImagePullBackOff   0          6m34s
pod/foo-deployment-6bbf665b47-dhndq   1/1     Running            0          40m
pod/foo-deployment-6bbf665b47-pnjfs   1/1     Running            0          40m

NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   23d

NAME                             READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/bar-deployment   3/3     1            3           7m6s
deployment.apps/foo-deployment   2/2     2            2           40m

NAME                                        DESIRED   CURRENT   READY   AGE
replicaset.apps/bar-deployment-75bcfbd655   3         3         3       7m6s
replicaset.apps/bar-deployment-864b65d8b6   1         1         0       6m34s
replicaset.apps/foo-deployment-6bbf665b47   2         2         2       40m
```

這時我們會發現我們的 `pod/bar-deployment` 發生了 `ImagePullBackOff` ，原因是我們並沒有建立

`mikehsu0618/bar:v1` 的 image ，這種情況很好的提供我們一個因為 `推進到一個不穩定的版本` 而需要使用 `版本回滾` 先復原服務到上一個正常的版本。

使用 `rollout` 的回滾指令復原先前版本設定：

```bash
// 回滾至上個版本
kubectl rollout undo deployment bar-deployment --record

// 回滾至指定版本
kubectl rollout undo deployment bar-deployment --to-revision=1 --record

----------------------------
deployment.apps/bar-deployment rolled back
```

這時 Deployment 已經回到了，沒有出問題的 `revision=1` 版本了

```bash
kubectl get all
----------------------------
NAME                                  READY   STATUS    RESTARTS   AGE
pod/bar-deployment-75bcfbd655-5b9z5   1/1     Running   0          17m
pod/bar-deployment-75bcfbd655-6whzr   1/1     Running   0          17m
pod/bar-deployment-75bcfbd655-zk88d   1/1     Running   0          17m
pod/foo-deployment-6bbf665b47-dhndq   1/1     Running   0          50m
pod/foo-deployment-6bbf665b47-pnjfs   1/1     Running   0          50m

NAME                 TYPE        CLUSTER-IP   EXTERNAL-IP   PORT(S)   AGE
service/kubernetes   ClusterIP   10.96.0.1    <none>        443/TCP   23d

NAME                             READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/bar-deployment   3/3     3            3           17m
deployment.apps/foo-deployment   2/2     2            2           50m

NAME                                        DESIRED   CURRENT   READY   AGE
replicaset.apps/bar-deployment-75bcfbd655   3         3         3       17m
replicaset.apps/bar-deployment-864b65d8b6   0         0         0       16m
replicaset.apps/foo-deployment-6bbf665b47   2         2         2       50m
```

## 結論

我們上面大致練習了幾個比較實用的方式，可以發現 `Deployment` 的設計非常的彈性以及簡潔，並且讓我們能將 `Pod` 設定在一起，大大的減少設定檔的數量。而 `Deployment` 因為可以簡單的設定 `水平擴展` `資源限制與請求` 等操作，使得許多進階觀念 `藍綠佈署` `金絲雀佈署` 得以更有可能的被一般的後端工程師實現（真是謝天謝地 wwww）。

Reference

**[Kubernetes 教學系列 - 滾動更新就用 Deployment](https://blog.kennycoder.io/2021/01/09/Kubernetes%E6%95%99%E5%AD%B8%E7%B3%BB%E5%88%97-%E6%BB%BE%E5%8B%95%E6%9B%B4%E6%96%B0%E5%B0%B1%E7%94%A8Deployment/)**

[雲原生社區-****Deployment****](https://jimmysong.io/kubernetes-handbook/concepts/deployment.html)

[Kubernetes Documentation-Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-back-a-deployment)

**[[Kubernetes] Deployment Overview](https://godleon.github.io/blog/Kubernetes/k8s-Deployment-Overview/)**

****[Kubernetes 基礎教學（二）實作範例：Pod、Service、Deployment、Ingress](https://cwhu.medium.com/kubernetes-implement-ingress-deployment-tutorial-7431c5f96c3e)****