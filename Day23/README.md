從異世界歸來的第二三天 - Kubernetes Resources(三) - LimitRange
---

## 概述

前兩天介紹完了 `Request/Limit` 以及 `Namespace` ，聰明的小朋友很快就會對這兩個觀念有所連結，一個能聲明單一服務的資源限制而另一個可以以資源規範在其之內的資源。

在默認情況下， `Kubernetes` 在容器的運行上使用的資源是沒有受到限制的，而身為管理者的我們可以利用 `Namespace` 為單位限制資源的使用以及創建，接下來我們就要繼續深入關於資源的進階觀念 `LimitRange` 。

### 什麼是 `LimitRange` ?
![https://ithelp.ithome.com.tw/upload/images/20220923/20149562hgaMUTpjSU.png](https://ithelp.ithome.com.tw/upload/images/20220923/20149562hgaMUTpjSU.png)
依附在 `Namespace` 下的 `LimitRange` 可以使管理員限制該 `Namespace` 下一個 `Pod` 或資源最多能夠使用資源配額所定義的 CPU 或內存用量。畢竟如果當一個 `Pod` 沒有被限制時，是有機會壟斷整個節點的可用資源的，而 `LimitRange` 即是在 `Namespace` 內限制資源分配的的策略對象(Policy)。

一個 **`LimitRange`** 提供的限制能夠做到：

- 在一個`Namespace`中實施對每個Pod 或Container 最小和最大的資源使用量的限制。
- 在一個`Namespace`中實施對每個PersistentVolumeClaim 能申請的最小和最大的存儲空間大小的限制。
- 在一個`Namespace`中實施對一種資源的申請值和限制值的比值的控制。
- 設置一個`Namespace`中對計算資源的默認申請/限制值，並且自動的在運行時注入到多個Container 中。

說那麼多接下來就來實際操作看看吧～

### 創建 Namespace

```jsx
kubectl create ns demo-namespace
```

將先建立好的 `Namespace` 設定為預設：

```jsx
kubectl config set-context --current --namespace=demo-namespace
--------
Context "docker-desktop" modified
```

### 創建聲明默認資源配額的 LimitRange

```jsx
# limit-range.yaml
apiVersion: v1
kind: LimitRange
metadata:
  name: limit-range
spec:
  limits:
		- max:
	      cpu: 1000m
				memory: 500Mi
	    min:
	      cpu: 500m
				memory: 200Mi 
      type: Container

```

查看一下剛剛建立的 `LimitRange` :

```jsx
kubectl get limitrange limit-range --output=yaml
-------
.....
limits:
  - default:
      cpu: "1"
      memory: 500Mi
    defaultRequest:
      cpu: 500m
      memory: 200Mi
    type: Container
```

建立了 `LimitRange` 後，在此 `Namespace` 下創建出 `Pod` 時，如果沒有特別聲明自己的資源配置， `Kubernetes` 就會依照 `LimitRange` 替該資源配額提供默認配置。如果該資源已有設定資源配額，而 `Kubernetes` 將會阻止超出規範的資源建立。

- 如 `Pod` 內的任何容器沒有聲明自己的請求以及限制，即為該容器設置默認的 CPU 和內存請求或限制。
- 確保每個 `Pod` 中的容器聲明的請求至少大於等於 `limits.defaultReqeust`
- 確保每個 `Pod` 中的容器聲明的請求至少小於等於  `limits.default`

### 創建一個沒有聲明請求和限制的 Pod

```jsx
# limit-range-pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: limit-range-pod
spec:
  containers:
  - name: default-limit-range-pod
    image: nginx
```

建立起資源：

```jsx
kubectl apply -f ./limit-range-pod.yaml
-------
pod/limit-range-pod created
```

查看 `LimitRange` 是否替我們配置了請求與限制：

```jsx
kubectl get pod limit-range-pod --output=yaml
--------
....
containers:
  - image: nginx
    imagePullPolicy: Always
    name: limit-range-pod
    resources:
      limits:
        cpu: "1"
        memory: 500Mi
      requests:
        cpu: 500m
        memory: 200Mi
....
```

成功看到相關配置～

### 當創建一個超過最大限制或不滿足最小請求的 Pod

當我們嘗試著建立一個超過 `LimitRange` CPU limit 的資源時， `Kubernetes` 將會直接返回以下類似錯誤訊息，因為其中定義了過高的 CPU limit :

```jsx
Error from server (Forbidden): error when creating "examples/admin/resource/limit-range-pod.yaml":
pods "limit-range-pod" is forbidden: maximum cpu usage per Container is 800m, but limit is 1500m.
```

反之當我們在嘗試著建立一個不滿足 `LimitRange` CPU request 的資源時，也會看到建立失敗的錯誤訊息：

```jsx
Error from server (Forbidden): error when creating "examples/admin/resource/limit-range-pod.yaml":
pods "limit-range-pod" is forbidden: minimum cpu usage per Container is 200m, but request is 100m.
```

## 結論

`Namespace` 使我們容易的做資源作分配，在加上 `LimitRange` 以及 `RequestQaota` 等以命名空間為單位的資源配置對象，讓我們可以靈活的位不同的部門規劃對應的資源配置。關於後者 `RequestQaota` 的使用方法一樣也是圍繞在 `Limit/Request` 上，相信這方面的觀念是一通百通，就稍微節省了一些篇幅讓各位細細的品嚐吧～ XD


相關文章：
- [從異世界歸來的第二一天 - Kubernetes Resources(一) - Request/Limit](https://ithelp.ithome.com.tw/articles/10295419)
- [從異世界歸來的第二二天 - Kubernetes Resources(二) - Namespace](https://ithelp.ithome.com.tw/articles/10296200)

相關程式碼同時收錄在：

[https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day23](https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day23)

Reference

****[為命名空間配置默認的CPU 請求和限制](https://kubernetes.io/zh-cn/docs/tasks/administer-cluster/manage-resources/cpu-default-namespace/#%E9%9B%86%E7%BE%A4%E7%AE%A1%E7%90%86%E5%91%98%E5%8F%82%E8%80%83)****

****[配置命名空间的最小和最大内存约束](https://kubernetes.io/zh-cn/docs/tasks/administer-cluster/manage-resources/memory-constraint-namespace/)****

****[為命名空間配置內存和CPU 配額](https://kubernetes.io/zh-cn/docs/tasks/administer-cluster/manage-resources/quota-memory-cpu-namespace/)****