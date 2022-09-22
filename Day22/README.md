從異世界歸來的第二二天 - Kubernetes Resources(二) - Namespace
---
## 概述

在昨天的介紹中我們藉由 `Request/Limit` 當作我們了解 `Kubernetes` 資源配置的一塊入門磚，今天我們將做一些實戰操作模擬工作中團隊開發的實際情況，首先我們需要介紹 　`Kubernetes` 為我們提供的一種集群中資源劃分並互相隔離的 Group - `Namespaces` ，並且在 `Namespace` 的下設置我們的資源配置。

### Namespace 是什麼以及何時使用？

![https://ithelp.ithome.com.tw/upload/images/20220922/20149562fG3VxjByFO.png](https://ithelp.ithome.com.tw/upload/images/20220922/20149562fG3VxjByFO.png)

Kubernetes 提供了`抽象的 Cluster (virtual cluster)`的概念，讓我們能根據專案不同、執行團隊不同，或是商業考量，將原本擁有實體資源的單一 Kubernetes Cluster ，劃分成幾個不同的`抽象的 Cluster (virtual cluster)`，也就是 `Namespace`。

所以他適用於存在很多跨團隊或者是項目的場景，對於只有少數幾個到十幾個使用者的集群，或許根本不需要創建或使用 `Namespace` 。套用一句知名前端框架 `Vue` 對 `Vuex` 下的一句精闢見解：『就像眼鏡一樣，你總會在需要他的時候想起他』。

### 查看`Namespace`

```jsx
kubectl get namespace
--------
NAME                   STATUS   AGE
default                Active   36h
kube-node-lease        Active   36h
kube-public            Active   36h
kube-system            Active   36h
kubernetes-dashboard   Active   36h
```

Kubernetes 會創建四個初始名字空間：

- `default` 沒有指明使用其它名字空間的對象所使用的默認名字空間
- `kube-system` ：Kubernetes 系統創建對象所使用的名字空間
- `kube-public` ：這個名字空間是自動創建的，所有用戶（包括未經過身份驗證的用戶）都可以讀取它。這個名字空間主要用於集群使用，以防某些資源在整個集群中應該是可見和可讀的。這個名字空間的公共方面只是一種約定，而不是要求。
- `kube-node-lease` ：該命名空間含有與每個節點關聯的Lease 對象。節點租用允許kubelet 發送heartbeat（心跳），以便控制平面能檢測節點故障。

<aside>
? 相信大家沒事都不會想去碰 `Kubernetes` 預設服務吧，雖然將其刪除時 `Kubernetes` 會竭盡全力的將服務重啟，但如果刪除到一半或中途出錯一定會很歡樂歐 ^__^。

</aside>

### 建立 `Namespace`

```jsx
kubectl create namespace demo-namespace
---------
namespace/demo-namespace creatednamespace     

                                                                        
kubectl get namespace
---------
NAME                   STATUS   AGE      
default                Active   37h
demo-namespace         Active   7s
kube-node-lease        Active   37h
kube-public            Active   37h
kube-system            Active   37h
kubernetes-dashboard   Active   36h
```

### 在**請求中設置**`Namespace`

要為當前請求設置名字空間，請使用 `--namespace` 參數。

例如：

```jsx
kubectl run nginx --image=nginx --namespace=demo-namespace
kubectl get pods --namespace=demo-namespace
```

### **設置預設** `Namespace`

在我們日常的 kubectl 指令中，使用資源的預設 `Namespace` 都是 `default` ，如果想要取得其他 `Namespace` 資源需要使用 `—namespace=<namespace>` 參數，我們還有另一個選擇就是修改預設的`Namespace`，以用於對應上下文中所有後續 kubectl 指令。

```jsx
kubectl config set-context --current --namespace=demo-namespace
*# 驗證*
kubectl config view --minify | grep namespace:
```

### 為 `Pod` 指定 `Namespace`

在我們纂寫 Pod 的設定檔中可以在 `[kubernetes.io/metadata.namespace](http://kubernetes.io/metadata.name)` 欄位指定要其運行在哪一個 `Namespace` 中，如果沒有特別設定將會視預設值而定。

```jsx
...
kind: pod
metatdata:
  namespace: <ns-name>
  name: <pod-name>
```

### 一些 `Namespace` 的特性

- 同一個 `Namespace` 的資源名稱是唯一性。
- 不同 `Namespace` 的資源名稱可以相同。
- `Namespace` delete 掉，裡面的 resources 也跟著刪除。
- 可透過 `ResourceQuota` `LimitRange` 分配/限制系統的資源。

相關程式碼同時收錄在：

[https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day22](https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day22)

Reference

****[为命名空间配置内存和 CPU 配额](https://kubernetes.io/zh-cn/docs/tasks/administer-cluster/manage-resources/quota-memory-cpu-namespace/)****

****[为命名空间配置默认的内存请求和限制](https://kubernetes.io/zh-cn/docs/tasks/administer-cluster/manage-resources/memory-default-namespace/)****

****[Kubernetes namespace 簡單介紹](https://sean22492249.medium.com/kubernetes-namespace-%E7%B0%A1%E5%96%AE%E4%BB%8B%E7%B4%B9-c48386949844)****

****[Namespaces](https://kubernetes.io/docs/concepts/overview/working-with-objects/namespaces/)****