# 從異世界歸來的第十一天 - Kubernetes Kubectl 指令與他的快樂夥伴

## 概述

在我們先前的介紹中，可以發現 `Kubernetes` 除了撰寫設定檔之外，其他時候就是使用 `kubectl` 這個指令工具進行各項操作，而 `kubectl` 的指令就是我們先前提到的 `kube-apiserver` 的這個元件以 `Restful API` 在底層各種調用的封裝，熟悉指令可以使我們更了解 `Kubernetes` 各種功能的觀念甚至幫助我們舉一反三調試出更靈活的設定。

## kubectl 語法介紹

如果想要熟悉 kubectl 語法更應該知道每個指令的用途以及語法格式，到後面就越來越可以利用累積的相關知識盲猜推測出更多指令操作，達到真正的內化。

`kubectl` 的語法如下：

```jsx
kubectl [command] [type] [name] [flags]
```

## **語法**

使用以下語法從終端窗口運行 `kubectl` 命令：

```jsx
kubectl [command] [TYPE] [NAME] [flags]
```

其中`command`、`TYPE`、`NAME`和 `flags` 分別是：

- `command`：指定要對一個或多個資源執行的操作，例如`create`、`get`、`describe`、`delete`。
- `TYPE`：指定[資源類型](https://kubernetes.io/zh-cn/docs/reference/kubectl/#resource-types)。資源類型不區分大小寫， 可以指定單數、複數或縮寫形式。例如，以下命令輸出相同的結果：
    1. 取得 pod：

        ```
        kubectl get pod pod1
        kubectl get pods pod1
        kubectl get po pod1
        
        ```

    2. 取得 service：

        ```
        kubectl get service service1
        kubectl get services service1
        kubectl get svc service1
        
        ```

    3. 取得 Deployment：

        ```
        kubectl get deployment deployment1
        kubectl get deployments deployment1
        kubectl get deploy deploymen1
        ```

- `NAME`：指定資源的名稱。名稱區分大小寫。如果省略名稱，則顯示所有資源的詳細信息。例如：`kubectl get pods`。

  在對多個資源執行操作時，你可以按類型和名稱指定每個資源，或指定一個或多個文件：

- 要按類型和名稱指定資源：
- 要對所有類型相同的資源進行分組，請執行以下操作：`TYPE1 name1 name2 name<#>`。

  例子：`kubectl get pod example-pod1 example-pod2`

- 分別指定多個資源類型：`TYPE1/name1 TYPE1/name2 TYPE2/name3 TYPE<#>/name<#>`。

  例子：`kubectl get pod/example-pod1 replicationcontroller/example-rc1`

- 用一個或多個文件指定資源：`f file1 -f file2 -f file<#>`
- [使用YAML 而不是JSON](https://kubernetes.io/zh-cn/docs/concepts/configuration/overview/#general-configuration-tips)， 因為YAML 對用戶更友好, 特別是對於配置文件。

  例子：`kubectl get -f ./pod.yaml`

- `flags`： 指定可選的參數。例如，可以使用 `-s` 或 `--server` 參數指定Kubernetes API 服務器的地址和端口。

<aside>
💡 **注意：從命令行指定的參數會覆蓋默認值和任何相應的環境變量。**

</aside>

如果你需要幫助，在終端窗口中運行`kubectl help`。

## 基礎語法

`**apply**` ：以文件或標準輸入為準應用或更新資源。

```jsx
# 使用 example-service.yaml 創建服務
kubectl apply -f example-service.yaml

# 使用 <directory> 路徑下的任意 .yaml、.yml 或 .json 文件 創建對象
kubectl apply -f <directory>
```

`**describe**` ：顯示一個或多個資源的詳細狀態，默認情況下包括未初始化的資源。

```jsx
# 顯示名為 <node-name> 的 Node 的詳細信息。
kubectl describe nodes <node-name>

# 顯示名為 <pod-name> 的 Pod 的詳細信息。
kubectl describe pods/<pod-name>

# 顯示由名為 <rc-name> 的副本控制器管理的所有 Pod 的詳细信息。
# 記住：副本控制器創建的任何 Pod 都以副本控制器的名稱为前缀。
kubectl describe pods <rc-name>

# 描述所有的 Pod
kubectl describe pods
```

**`get` ：**用於獲取集群的一個或一些resource信息。

該命令可以列出集群所有資源的詳細信息，resource包括集群節點、運行的Pod、Deployment、Service等。

<aside>
💡 集群中可以創建多個namespace，未指定namespace的情況下，所有操作都是針對--namespace=default。

</aside>

例如：

獲取所有pod的詳細信息：

```
kubectl get po -o wide
```

獲取所有namespace下的運行的所有pod：

```
kubectl get po --all-namespaces
```

獲取所有namespace下的運行的所有pod的標籤：

```
kubectl get po --show-labels
```

獲取該節點的所有命名空間：

```
kubectl get namespace
```

<aside>
💡 查詢其他節點需要加-s指定節點，類似可以使用“kubectl get svc”，“kubectl get nodes”，“kubectl get deploy”等獲取其他resource信息。

</aside>

**`create`** ：根據文件或者輸入來創建資源。

```jsx
kubectl create -f demo-deployment.yaml
kubectl create -f demo-service.yaml
```

**`delete`** ：刪除資源。

```jsx
kubectl delete -f demo-deployment.yaml
kubectl delete -f demo-service.yaml
kubectl delete {具體資源的名稱}
```

`**run`** ：在集群中創建並運行一個或多個[容器鏡像](https://cloud.tencent.com/product/tcr?from=10680)。

```jsx
// 語法
kubectl run NAME --image=image [--env="key=value"] ＼
	[--port=port] [--replicas=replicas] [--dry-run=bool] ＼
	[--overrides=inline-json] [--command] -- [COMMAND] [args...]
```

```jsx
// 運行一個名稱為nginx，副本数為3，標籤為app=example，鏡像為nginx:1.10，端口為80的容器實例
kubectl run nginx --replicas=3 --labels="app=example" --image=nginx:1.10 --port=80
```

**`expose`** ：創建一個service服務，並且暴露端口讓外部可以訪問。

```jsx
# 創建一個 nginx 服務並暴露 88 端口讓外部訪問
kubectl expose deployment nginx --port=88 --type=NodePort --target-port=80 --name=nginx-service

```

**`set`** ：配置應用的一些特定資源，也可以修改應用已有的資源。

```jsx
//使用 kubectl set --help查看，它的子命令，env，image，resources，selector，serviceaccount，subject。

// 語法
kubectl resources (-f FILENAME | TYPE NAME) ([--limits=LIMITS & --requests=REQUESTS]
```

`**exec**` ：對Pod 中的容器執行命令。

```jsx
# 從 Pod <pod-name> 中獲取運行 'date' 的輸出。默認情况下，輸出来自第一個容器。
kubectl exec <pod-name> -- date

# 運行輸出 'date' 獲取在 Pod <pod-name> 中容器 <container-name> 的輸出。
kubectl exec <pod-name> -c <container-name> -- date

# 獲取一个交互 TTY 並在 Pod  <pod-name> 中運行 /bin/bash。默認情况下，輸出来自第一個容器。
kubectl exec -ti <pod-name> -- /bin/bash
```

`**logs**` ：打印Pod 中容器的日誌。

```jsx
# 返回 Pod <pod-name> 的日誌快照。
kubectl logs <pod-name>

# 從 Pod <pod-name> 開始流式傳輸日誌。這種類似於 'tail -f' Linux 命令。
kubectl logs -f <pod-name>
```

## 概述

到目前為止我們已經了解了基本的指令以及設定，之後的方向我們將逐漸揭開 `Kubernetes` 的神秘面紗，使用更實際以及更深入的例子來學習各種設定，以及了解該設定的存在的原因。

Reference

**[Kubernetes 教學系列 - kubectl 常見指令說明](https://blog.kennycoder.io/2020/12/18/Kubernetes%E6%95%99%E5%AD%B8%E7%B3%BB%E5%88%97-kubectl%E5%B8%B8%E8%A6%8B%E6%8C%87%E4%BB%A4%E8%AA%AA%E6%98%8E/)**

****[kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)****

****[Command line tool (kubectl)](https://kubernetes.io/docs/reference/kubectl/)****