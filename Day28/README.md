從異世界歸來的第二八天 - Kubernetes Authorization (一) - 使用 Context 進行用戶管理
---

## 概述

結束完了 `AutoScaling` 主題後，隨即而來的是本系列的最後一個主題 `Security` ，這個議題可以說在 DevOps 中是特別耐人尋味的，有趣的點是他不一定是工作中的必備技能但卻極為重要，畢竟不是每個人公司都有數十人等級以上的團隊、或同時管理維護大大小小不同產品的工作環境，更多的是三到五人的小開發團隊以及一人維護的專案的情況，這時候權限管理相關的優先級自然被排到比較後面。

這時我們從另一個角度來看，人人想進的理想好公司通常除了好的文化跟環境之外，也有很大的概率是一間人數不小規模的企業甚至是跨國集團，如果我們在職業生涯追求的是這類型的公司，那又有什麼理由不去理解為自己加分的技能呢？而我想在資安就是一個不論是否為維運人員都很適合熟悉的一環。

## Kubernetes 的認證與授權

當我們試著不管藉由指令或是其他方式與我們的 `Kubernetes` 集群互動時，一定都會經過該集群管理所有物件資源的的入口 `kube-apiserver` ，這時就可以體現 `kube-apiserver` 身份驗證的重要性，他能替我們檢查所有進入的請求，包括 Authentication(身份驗證)、Authorization(授權)和 Admission Control(准入控制)。

Kubernetes API 請求從發起到持久化到ETCD資料庫中的過程如下：

![https://ithelp.ithome.com.tw/upload/images/20220928/20149562xWEHihpdXJ.png](https://ithelp.ithome.com.tw/upload/images/20220928/20149562xWEHihpdXJ.png)

接下來我們使用 Kubernetes Context 實戰我們在工作中會面對到的用戶管理認證(Authentication) 問題。

## Kubernetes Context 是什麼？

![https://ithelp.ithome.com.tw/upload/images/20220928/20149562DuzrV22d16.png](https://ithelp.ithome.com.tw/upload/images/20220928/20149562DuzrV22d16.png)

在 `Kubernetes` 中，一個 Context 就像是一個方便好讀在客戶端紀錄著你與集群要如何溝通的 `Alias` ，假如我切換到某個群集後，那我在 kubectl 送出的每個指令都會指向該 `Context` 所設定的 `cluster` 、`namespace` 、`user` 執行操作。很重要的一點是 `kube-apiserver` 是看不懂所謂的 `Context` 的，而是在請求送出之前幫我們在客戶端把相關設定轉為參數送出，所以才呼應了前面提到的 `Alias` 概念。

先來看看 kubeconfig 中紀錄的 Context 到底長什麼樣吧：

```yaml
kubectl config view
---------

apiVersion: v1ig view               
clusters:
- cluster:
    certificate-authority-data: DATA+OMITTED
    server: https://kubernetes.docker.internal:6443
  name: docker-desktop
contexts:
- context:
    cluster: docker-desktop
    user: docker-desktop
  name: docker-desktop
current-context: docker-desktop
kind: Config
preferences: {}
users:
- name: docker-desktop
  user:
    client-certificate-data: REDACTED
    client-key-data: REDACTED
```

只查看當前的 context config 資訊：

```yaml
kubectl config view --minify
```

基本上可以歸納出三個重點：

- **`Clusters`** ：此列表中紀錄著每個集群如何與該集群的 `kube-apiserver` 溝通的 URL以及認證權限。
- **`Contexts`** ：此列表中紀錄著當你在該 `Context` 時，對 `kube-apiserver` 溝通時將會以內容中的 `cluster` 、`user` 、`namespace` 來執行（如 namespcae 為空則預設為 `default` ）。
- **`Users`** ：使用者選項定義著每個使用著唯一的名字以及相關的多種認證授權資訊，像是 `client certificates` 、`bearer tokens` 、`authenticating proxy` 等等。

## 用戶管理情境

假設在我們的工作環境中有兩個集群，一個用於正式環境(production)，另一個用於開發環境(develop)，而兩個集群中我們又將前端(frontend)和後端(backend)用 `namespace`區隔開來，接著我們將創造出前端開發者以及後端開發者兩種角色，並期望使用 `Context` 規範不管在正式環境或開發環境中，前端開發者與後端開發者只能在對應的 `namespace` 下控制集群。

此時粗略的 kubeconfig 大概如下：

```yaml
apiVersion: v1
kind: Config

clusters:
- cluster:
  name: development
- cluster:
  name: production

users:
- name: backend-developer
- name: front-developer

contexts:
- context:
  name: dev-backend
		namespace: backend
		cluster: development
		user: backend-developer
- context:
  name: prod-backend
		namespace: backend
		cluster: production
		user: backend-developer
- context:
  name: dev-frontend
		namespace: frontend
		cluster: development
		user: frontend-developer
- context:
  name: prod-frontend
		namespace: frontend
		cluster: production
		user: frontend-developer
```

可以簡單的理解成，當切換到哪個 `Context` 時，即可以視為該 `user` 認證身份下的『操作權限』在該 `cluster` 中的`kube-apiserver` 發出對該 `namespace`  下的資源請求。

### 實際操作

1. 首先建立出所有的 `Context` ：

```yaml

# 建立/修改 Context 指令：
kubectl config set-context <CONTEXT_NAME> --namespace=<NAMESPACE_NAME>--cluster=<CLUSTER_NAME> --user=<USER_NAME>

# 陸續將所有 Context 建立出來：
kubectl config set-context prod-frontend --cluster=production --namespace=frontend --user=prod-frontend
--------
Context "prod-frontend" created.

kubectl config set-context prod-backend --cluster=production --namespace=backend --user=prod-backend
--------
Context "prod-backend" created.

kubectl config set-context dev-frontend --cluster=development --namespace=frontend --user=dev-frontend
--------
Context "dev-frontend" created.

kubectl config set-context dev-backend --cluster=development --namespace=backend --user=dev-backend
--------
Context "dev-backend" created.
```

再次查看 kubeconfig 可以發現上列 `Context` 以出現在列表之中：

```yaml
kubectl config view
-------
contexts:
- context:
    cluster: development
    namespace: backend
    user: dev-backend
  name: dev-backend
- context:
    cluster: development
    namespace: frontend
    user: dev-frontend
  name: dev-frontend
- context:
    cluster: docker-desktop
    user: docker-desktop
  name: docker-desktop
- context:
    cluster: production
    namespace: backend
    user: prod-backend
  name: prod-backend
- context:
    cluster: production
    namespace: frontend
    user: prod-frontend
  name: prod-frontend
```

1. 切換到所需的 `Context` ：

```yaml
# 顯示當前使用的 context
kubectl config current-context

# 切換到指定的 context 
kubectl config use-context <CONTEXT_NAME>
```

此時我們已經可以隨心所欲的設定並切換 `Context` ，在團隊中只要讓對應的角色設定到正確的 `Context` ，即可實現讓該 `user` 在指定 `cluster` 中執行被賦予的權限。

1. 刪除 config 資源：

```yaml
# 刪除用戶
kubectl --kubeconfig=config-demo config unset users.<name>
# 刪除集群
kubectl --kubeconfig=config-demo config unset clusters.<name>
# 刪除 context
kubectl --kubeconfig=config-demo config unset contexts.<name>
```

### 所以說那個 Context 中的 Cluster 跟 User 呢？

看到這裡可能有些敏銳的同學注意到了，既然 `Context` 如前面所說的是一種 `Alias` 的概念，那實際執行的 `cluster` 授權以及 `user` 認證是如何來的呢？

由於我們在本地使用的 docker-desktop 只能提供我們一個 `cluster` 集群，加上在實際工作中使用的大多是雲端平台整合好的 `Kubernetes` 服務（Google GKE, AWS EKS….等），各家平台對於 `kubeconfig` 的 `cluster`  管理多半有與自家用戶權限整合，並以自家 cli 或 sdk 的方式進行操作，各家操作方式不一，所以本篇重點將放在 Context 管理上。

簡單使用 Google 的 `gcloud auth` 相關指令新增了對 Google GKE 集群的 kubeconfig，大概會如下所示：

```yaml
kubectl config view
--------
apiVersion: v1                                                                                  
clusters:
- cluster:
    certificate-authority-data: DATA+OMITTED
    server: https://35.xxx.xxx.xxx
  name: gke_xxx-xxxx_asia-east1-c_xxx-dev
contexts:
- context:
    cluster: gke_xxx-xxx_asia-east1-c_xxx-dev
    namespace: rtmp-relay
    user: gke_xxx-xxx_asia-east1-c_xxx-dev
  name: gke_xxx-xxx_asia-east1-c_xxx-dev
current-context: docker-desktop
kind: Config
preferences: {}
users:
- name: gke_xxx-xxx_asia-east1-c_xxx-dev
  user:
    auth-provider:
      config:
        access-token: 
        cmd-args: config config-helper --format=json
        cmd-path: /google-cloud-sdk/bin/gcloud
        expiry: "2022-09-23T17:43:59Z"
        expiry-key: '{.credential.token_expiry}'
        token-key: '{.credential.access_token}'
      name: gcp
```

關於使用 kubectl 獨自建立/修改 `cluster` 的指令：

```yaml
kubectl config set-cluster \
<CLUSTER_NAME> \
--server=<SERVER_ADDRESS> \
--certificate-authority=<CLUSTER_CERTIFICATE>
```

而建立/修改 `user` 的指令也是非常相似的：

```yaml
kubectl config set-credentials \
<USER_NAME> \
--client-certificate=<USER_CERTIFICATE> \
--client-key=<USER_KEY>
```

以上指令使用憑證的方式驗證身份，而 `--client-certificate` 被用來認證 `user` ，`--certificate-authority` 則用來認證該 `cluster` 。官方提供更多的驗證方式可以參考相關文件 [Kubernetes authentication overview](https://kubernetes.io/docs/reference/access-authn-authz/authentication/)。

## 結論

在以上的介紹中，我們大概可以了解 `Context` 這個在很多語言或者工具中，有著不同用途的抽象概念。有了 `Context` 我們在身為服務守護者的團隊協作中，可以確保每個人該有的最大權限，在個人開發中，也可以防止自己的人為疏忽而造成不可逆轉的損失，接下來我們將更進一步的深入介紹 ，基於`RBAC` 下，我們就來看看 `Kubernetes` 如何對一個角色進行授權吧。

相關程式碼同時收錄在：

[https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day28](https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day28)

Reference

****[配置對多集群的訪問](https://kubernetes.io/zh-cn/docs/tasks/access-application-cluster/configure-access-multiple-clusters/)****

****[基於角色的訪問控制（RBAC）](https://jimmysong.io/kubernetes-handbook/concepts/rbac.html)****

**[了解 Kubernetes 中的認證機制](https://godleon.github.io/blog/Kubernetes/k8s-API-Authentication/)**

****[管理服務賬號](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/service-accounts-admin/)****

****[Kubectl Config Set-Context](https://www.containiq.com/post/kubectl-config-set-context-tutorial-and-best-practices)****

****[k8s 基於RBAC的認證、授權介紹和實踐](https://iter01.com/657295.html)****

****[Day 19 - 老闆！我可以做什麼：RBAC](https://ithelp.ithome.com.tw/articles/10195944)****

****[[Day17] k8s管理篇（三）：User Management、RBAC、Node Ｍaintenance](https://ithelp.ithome.com.tw/articles/10223717)****

****[用戶認證](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/authentication/)****