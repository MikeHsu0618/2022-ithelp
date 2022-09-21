從異世界歸來的第二一天 - Kubernetes Resources(一) - Request/Limit
---

## 概述

在我們大略介紹了常見的存儲配置 `Volumes` ，接下來我們將會慢慢的進入資源分配的世界。 `Kubernetes` 是一個集群管理平台並且擁有至少一個節點，每個節點都可以被理解成一台物理主機，所以 `Kubernetes` 理所當然的需要統計整個平台的資源使用情況，合理的將資源分配給容器使用，並且保證容器生命週期內有足夠的資源來保證其可運作。反面來說，如果設定不當使得資源被某些空閒的容器獨占著不放反而是非常浪費的。

`Kubernetes` 需要考慮如何在優先度和公平性的前提下提供資源的利用率，為了實現資源被有效調度和分配時同時提高資源利用率， `Kubernetes` 提供了 `Request/Limit` 兩種限制類型讓我們對資源進行分配。

## 什麼是 Request 以及 Limit ?

### request

- 容器使用的最小資源要求，做為容器調度時資源分配的判斷依賴。
- 只有當前節點上可分配的資源量 `>= request` 時才允許將容器調度到該節點上。

### limit

- 容器能使用的最大值。
- 設置為 0 表示對使用的資源不做限制，可以無限使用。

### **request 和limit 關係**

容器所聲明的 `request`應該大於等於0並且不超過節點的可分配容量。這條規則可用以下公式總結：

```jsx
0 <= request <= Node Allocatable
```

而 `limit` 則應該大於等於`request` 並且其值無上限：

```jsx
request <= limit <= Infinity
```

## Resource 的種類 ： CPU、Memory

![](https://478h5m1yrfsa3bbe262u7muv-wpengine.netdna-ssl.com/wp-content/uploads/image11.png)

Kubernetes將底層處理器架構抽象為了計算資源，將它們按照需求暴露為原始值或基本單位。

- CPU：對於CPU資源來說，這些基本單位是基於核心(cores)的；

  而一個CPU則相當於：

  - 一個AWS vCPU
  - 一個GCP Core
  - 一個Azure vCore
  - 英特爾處理器上一個 Hyperthread(處理器要支持Hyperthreading)
- Memory：對於內存來說，則是基於字節的。內存資源可以使用單純的數值或帶有後綴(E、P、T、G、M、K)的定點整數表示，也就是我們常見的單位。

## Pod 的服務品質（QoS aka. Quality of Service)

Kubernetes 創建Pod 時就會依照設定的`Request/Limit`給它指定了下列一種QoS 類：

### Guaranteed

當一個Pod內的每個容器，其 request.memory 等於 limit.memory 且 request.cpu 等於 limit.cpu時，這個Pod被認為是Guaranteed。

```jsx
apiVersion: v1
kind: Pod
metadata:
  name: qos-demo
  namespace: qos-example
spec:
  containers:
  - name: qos-demo-ctr
    image: nginx
    resources:
      limits:
        memory: "200Mi"
        cpu: "700m"
      requests:
        memory: "200Mi"
        cpu: "700m"
```

### Burstable

需要滿足2個條件:

- 不是Guaranteed Pod。
- Pod內至少有一個容器設置了memory 或cpu request。

```jsx
apiVersion: v1
kind: Pod
metadata:
  name: qos-demo-2
  namespace: qos-example
spec:
  containers:
  - name: qos-demo-2-ctr
    image: nginx
    resources:
      limits:
        memory: "200Mi"
      requests:
        memory: "100Mi"
```

### BestEffort

如果一個Pods內的所有容器都沒有設置`request`和`limit`，則這個 `pod` 被認為是 `best-effort`。

```jsx
apiVersion: v1
kind: Pod
metadata:
  name: qos-demo-3
  namespace: qos-example
spec:
  containers:
  - name: qos-demo-3-ctr
    image: nginx
```

查上以上產生出來的 `qosClass` **：**

```jsx
kubectl get pod qos-demo-3 --output=yaml
---------

spec:
  containers:
    ...
    resources: {}
  ...
status:
  qosClass: BestEffort / Burstable / Guaranteed
```

Kubernetes根據上述不同類型的pod，將給出不同的資源使用權和優先級。 `Best-Effort` Pods 有著最低的優先級，在系統內存不足時，它們是第一批被清理的對象。`Guaranteed` Pods有著最高的優先級，通常不會被殺死或節流，除非資源使用超過了limits的限制並且沒有其它更低優先級的pods可清理了。最後，`Burstable` Pods有著最小的資源保證但是條件允許時允許使用更多的計算資源。在沒有`Best-Effort` Pods存在並且系統容量不足時，`Burstable` Pods將是集群中第一批被殺死的。

## 結論

看完上述的觀念介紹我們可以了解到 `request` 和 `limit` 與 Pods 的命運息息相關，主要的動機不外乎是想要更安全高效的使用計算資源，並確保高優先級的 Pods 正常執行，同時保證資源不會被過度使用。此外如果將 `limit` 的設定高於 `request` ，則代表著當資源充足時，Pods 可以利用更多的資源。


相關程式碼同時收錄在：

[https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day21](https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day21)

Reference

****[配置Pod 的服務質量](https://kubernetes.io/zh-cn/docs/tasks/configure-pod-container/quality-service-pod/)****

****[Kubernetes資源分配(limit/request)](https://developer.aliyun.com/article/679986)****

****[【譯】Kubernetes中的資源分配](https://www.modb.pro/db/46091)****