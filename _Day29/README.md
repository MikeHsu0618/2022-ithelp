## 概述

在昨天的 Context 介紹之中，我們了解到在 `Kubernetes` 中，是如何管理用戶或者是群組，但是沒有對其的認證機制有太多著墨。那是因為在 `Kubernetes` 中的認證授權是一門相當大的學問，其中實現的方式跟概念非常複雜，這邊我們就來簡單談談與應用層形影不離的 `RBAC` 授權。

## 深入了解 Kubernetes API Server

在任何想要取得 `Kubernetes` 資源的一端，都需要經過 `Authentication`(身份認證) 以及 `Authorization` (授權)的驗證過程， 才能依照被授權的權限進行操作。

### Authentication 身份認證

簡單來說就是確認該使用者是否能夠進入 `Kubernetes` 集群中，一般來說可以分為兩種使用者 - 普通使用者(User) 以及服務帳號(Service Account)。

- 普通使用者(User)：可以簡單理解成，可以通過 kubectl 指令、或Restful 請求的身份認證，就可以視為一個普通使用者，而這也是我們設定在 `Context` 中的 `user` 。
- 服務帳號(Service Account)：`Service Account` 本身在 `Kubernetes` 是屬於 resource 的一種，與普通使用者的全局性不同的是，Service account 是以 namespace 為作用域單位。其針對執行中的 Pod 而言，每個 namespace 被建立時，`Kubernetes` 都會隨之建立一個名稱為 `default`的 `Service account` 並帶有 token 供未來在此 namespace 中產生的 Pod 使用，所以 Pod 將會依照該 `Service account` 的 token 與 API server 進行認證。

而在 `Kubernetes` 中有幾種驗證方式：

- Certificate
- Token
- OpenID
- Web Hook

其中 Certificate 是在普通使用者中被廣泛使用的驗證方式。通過客戶端證書進行身份驗證時，客戶端必須先取得一個有效的 `X.509` 客戶端證書，然後由 Kubernetes API Server 通過驗證這個證書來驗證你的身份。當然這個 `X.509` 證書必須由集群 CA 證書簽名，看起來有沒有跟 HTTPS 證書很相似，其中差別只在於 CA 供應方變成 `Kubernetes` 而已。

最後我們將會實作 `X.509` 證書完成身份認證，如果對 Certificate 還不夠熟悉的話，可以參考[這篇文章](https://hackmd.io/@yzai/rJXYxFpmq)。

### Authorization 授權

當我們通過了 `Authentication` (身份認證)後，那僅能代表當前的使用者允許與 Kubernetes API Server 溝通，至於該使用者是否有權限(Permission)請求什麼資源，就是 `Authorization` 該登場的時候了，Kubernetes API Server 將會在時候審查請求的 API 屬性，像是 user、Api request verb、Http request verb、Resource、Namespace、Api Group… 等。

關於 Authorization Mode 有以下幾種模式：

- Node
- ABAC
- RBAC
- Webhook

以下我們會以 `RBAC` 做個介紹並實戰練習。

此外`kubectl`提供 `auth can-i` 子命令，用於快速查詢 API 審查：

```yaml
# 檢查是否可以對 deployments 執行 create
kubectl auth can-i create deployments --namespace default
-----
yes
```

## 實戰使用 RBAC(Role-Base Access Control)

![https://ithelp.ithome.com.tw/upload/images/20220929/20149562O5YPYptWMy.png](https://ithelp.ithome.com.tw/upload/images/20220929/20149562O5YPYptWMy.png)

`Role-Base Access Control` 顧名思義是指基於 Role 的概念建立的訪問控制，用來調節使用者對 Kubernetes API Server 的訪問的方法，在各類大型系統以及各種雲端平台中廣泛使用。

`RBAC` 在我們實際工作環境中，就像是超級管理者這個角色一定跟瀏覽者擁有巨大的權限差距一樣，做為內部重要的雲端基礎建設並不希望每個使用者都可以不受限制的建立、刪除資源，為此我們實現了將權限(Permission) 綁定普通使用者以及服務帳號的概念，將複雜的業務權限做的更輕量化，並遵從了`權限最小化原則` 。

接下來我們將實現使用者取得認證並使用 `RBAC` 進行角色權限綁定：

### 1. 建立 Context 並以 X.509 證書驗證普通使用者（user）

首先我們要以自身 `Kubernetes` 做為 CA 發送方來產生 CA 證書。

產生一個使用者私鑰：

```yaml
openssl genrsa -out pod-viewer.key 2048
----
-----BEGIN RSA PRIVATE KEY-----

// ...

-----END RSA PRIVATE KEY-----
```

通過 pod-viewer.key 去產生 CSR (證書簽名請求)，`Kubernetes` 將會使用證書中的 'subject' 的通用名稱（Common Name）欄位來確定使用者名稱：

```yaml
openssl req -new -key pod-viewer.key -out pod-viewer.csr -subj "/CN=pod-viewer/O=app"
------
-----BEGIN CERTIFICATE REQUEST-----

// ... 

-----END CERTIFICATE REQUEST-----
```

有了CSR，我們就可以把它交給集群管理者(在這裡是指我們) 通過集群 CA 簽署客戶端證書。

所以我們需要去 `docker-desktop` 的節點中取得 CA 根憑證，用此來簽署所有須要通訊的證書。

下載 **[kubectl-node-shell](https://github.com/kvaps/kubectl-node-shell)** 進入 `docker-desktop` 節點並取得 CA 內容：

```yaml
curl -LO https://github.com/kvaps/kubectl-node-shell/raw/master/kubectl-node_shell
chmod +x ./kubectl-node_shell
sudo mv ./kubectl-node_shell /usr/local/bin/kubectl-node_shell

kubectl node-shell docker-desktop // 進入 docker-desktop 節點

cat /etc/kubernetes/pki/ca.crt // 在本地新增一個 ca.crt 檔案將打印出的內容填入
cat /etc/kubernetes/pki/ca.key // 在本地新增一個 ca.key 檔案將打印出的內容填入
```

此時我們的目錄底下應該會有以下四個檔案：

```yaml
ls
------
ca.crt            ca.key            pod-viewer.csr    pod-viewer.key
```

接下來就用拿到的根憑證來產生一個被集群 CA 簽署過的證書：

```yaml
openssl x509 -req -in pod-viewer.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out pod-viewer.crt -days 365
```

此時我們將會獲得一個 `pod-viewer.crt` 的證書，可以使用以下指令來查看詳細內容：

```yaml
openssl x509 -noout -text -in pod-viewer.crt
------
Data:
        Version: 1 (0x0)
        Serial Number:
            8f:ac:d9:57:79:80:11:8d
    Signature Algorithm: sha256WithRSAEncryption
        Issuer: CN=kubernetes
        Validity
            Not Before: Sep 27 10:24:39 2022 GMT
            Not After : Sep 27 10:24:39 2023 GMT
        Subject: CN=pod-viewer
        Subject Public Key Info:
            Public Key Algorithm: rsaEncryption
                RSA Public-Key: (2048 bit)
                Modulus:
// ...
```

到這裡我們有了證書之後就可以開始建立 `Context`與`User`了。

使用證書建立一個 `User` ：

```yaml
kubectl config set-credentials pod-viewer \
    --client-certificate=ca.crt \
    --client-key=ca.key \
    --embed-certs=true
-------
User "pod-viewer" set.
```

使用上一篇的指令建立一個 `Context` ：

```yaml
kubectl config set-context only-view --cluster=docker-desktop --user=pod-viewer
------
Context "only-view" created.
```

這裡我們建立了一個 `Context` ，指向我們現有的 `docker-desktop` 集群，而 `user` 是還沒被授權的 `pod-viewer` 。

查看 `kubeconfig` 中的設定一下：

```yaml
kubectl config view
------

apiVersion: v1
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
- context:
    cluster: docker-desktop
    user: pod-viewer
  name: only-view
current-context: docker-desktop
kind: Config
preferences: {}
users:
- name: docker-desktop
  user:
    client-certificate-data: REDACTED
    client-key-data: REDACTED
- name: pod-viewer
  user:
    client-certificate-data: REDACTED
    client-key-data: REDACTED
```

可以看到我們的 `Context` 已經配置完畢，但此時我們只完成了 `Authentication` ，並沒有獲得任何權限，可以大膽猜測目前我們可以與集群溝通但不能取得任何資源。

切換到 pod-viewer context 並試圖查看 pod 資訊：

```yaml
kubectl config use-context only-view
-------
Switched to context "only-view

kubectl get pod
-------
Error from server (Forbidden): pods is forbidden: User "pod-viewer" cannot list resource "pods" in API group "" in the namespace "default"
```

如預期的 `only-view` 的 pod-viewer 並沒有擁有任何權限，所以不能通過 `Authorization` 。

### 2. 使用 RBAC 授權給使用者

在開始操作之前我們需要對 `RBAC` 有進一步的了解，它是 `Kubernetes` v1.8 正式引入的 Authorization 機制，也就是一種管制訪問 k8s API 的機制。管理者可以透過 `rbac.authorization.k8s.io`這個 API 群組來進行動態的管理配置。主要由 Role、ClusterRole、RoleBinding、ClusterRoleBinding 等資源組成。

![https://ithelp.ithome.com.tw/upload/images/20220929/20149562PYHD14cXKE.png](https://ithelp.ithome.com.tw/upload/images/20220929/20149562PYHD14cXKE.png)

透過適當的教色配置與授權分配，管理者可以決定使用者可以使用哪些功能。在 `RBAC` 下的角色會被賦予指定的權限(permission) 並實現最小權限源則，對比於限制特定權限的方式更為嚴謹。

### Role vs ClusterRole

角色是一組許可規則的集合，`Role` 用來定義某個 `namespace` 內的訪問許可，而 `ClusterRole` 則是一個集群資源。有個以上兩種不同顆粒細度的資源作用範圍，可以使我們更好方便將權限拆分的更仔細。如果系統管理員沒有 `ClusterRole` 的權限，那代表的是他需要將每個 `namespace` 一一綁定到需要集群等級權限的使用者，那將是 DevOps 的一大夢魘。

就來看看實際例子：

```yaml
# role.yaml
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  namespace: default # 定義在 default 命名空間
  name: pod-viewer       # Role 名稱 
rules:
- apiGroups: [""] # “” 默認代表 apiVersion:v1
  resources: ["pods"]
  verbs: ["get", "watch", "list"]
```

`role.yaml` 聲明了一個作用在 namespace `kube-system` 中的 Role 物件，並允許 Role 能夠對 `pods` 進行限定操作。

構成一個 **Rule** 需要宣告三部分：

- `apiGroups`：資源所屬的API組：`""` 預設為 core 組資源，如：extensions、apps、batch等。[Kubernetes API 參考文件](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.23/#-strong-api-groups-strong-)
- `resources`：資源，如： pods、deployments、services、secrets 等。
- `verbs`：動作，如： get、list、watch、create、delete、update 等。

在提供一個 ClusterRole 的例子：

```yaml
# cluster-role.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: cluster-pod-viewer
rules:
- apiGroups: [""] 
  resources: ["pods"]
  verbs: ["get", "list", "watch"]
```

很清楚的看出兩者的差異在於 `ClusterRole` 並不需要指定 `namespace` 。

首先需要將 context 切回集群管理者的 context 才有權限建立起剛剛的 `Role` ：

```yaml
# 記得先切換回 docker-desktop
kubectl config use-context docker-desktop
------
Switched to context "docker-desktop".

kubectl apply -f role.yaml
------
role.rbac.authorization.k8s.io/pod-viewer created
```

成功建立！但這時我們還未將任何使用者與此 `Role` 資源綁定，所以接下來就要設定 `RoleBinding` 或 `ClusterRoleBinding` 。

### ****RoleBinding vs ClusterRoleBinding****

以上我們已經擁有了一個帶有授權的 Role，下一步我們需要將此角色綁定到指定使用者，才能將角色中定義好的授權賦予給一個或一組使用者使用，及是以下的 `Subject` 代表被綁定的對象。

![https://ithelp.ithome.com.tw/upload/images/20220929/20149562tW6IxaufUw.png](https://ithelp.ithome.com.tw/upload/images/20220929/20149562tW6IxaufUw.png)

被綁定的對象可以是

`User` ：對於名稱為 `alice@example.com` 的用戶

```yaml
**subjects**:
- **kind**: User
  **name**: "alice@example.com"
  **apiGroup**: rbac.authorization.k8s.io
```

`Service Account` ：對於 `kube-system` 命名空間中的 `default` 服務賬戶

```yaml
**subjects**:
- **kind**: ServiceAccount
  **name**: default
  **namespace**: kube-system
```

`Group` ：在 `Kubernetes` 中我們可以指定符合特定前綴，將符合條件的使用者劃分為同一組。

```yaml
# 對於"qa" 名稱空間中的所有服務賬戶
subjects:
- kind: Group
  name: system:serviceaccounts:qa
  apiGroup: rbac.authorization.k8s.io
# 對於所有用戶
subjects:
- kind: Group
  name: system:authenticated
  apiGroup: rbac.authorization.k8s.io
- kind: Group
  name: system:unauthenticated
  apiGroup: rbac.authorization.k8s.io
```

<aside>
? 前綴 `system:` 是Kubernetes 系統保留的，所以你要確保所配置的用戶名或者組名不能出現上述 `system:` 前綴。除了對前綴的限制之外，RBAC 鑑權系統不對用戶名格式作任何要求。

</aside>

接下來就讓我們將 `RoleBinding` 實現出來：

```yaml
# role-binding.yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: pod-viewer-rolebinding
  namespace: default #授權的名稱空間為 default
subjects:
  - kind: User
    name: pod-viewer # 繫結 pod-viewer 使用者
    apiGroup: rbac.authorization.k8s.io
roleRef:
  kind: Role
  name: pod-viewer # 繫結 Role
  apiGroup: rbac.authorization.k8s.io
```

建立 `RoleBinding` 資源：

```yaml
kubectl apply -f role-binding.yaml
------
rolebinding.rbac.authorization.k8s.io/pod-viewer-rolebinding created
```

這時就可以切回我們先前建立好的 `pod-viewer` 的 context：

```yaml
kubectl config use-context pod-viewer
------
Switched to context "pod-viewer".
```

現在就來驗證相關權限是否如預期綁定：

```yaml
# 成功取得 pod 資訊(此時沒有任何 pod 在執行)
  kubectl get pod -n default
  ------
  No resources found in default namespace.

  # 成功收到 forbidden 阻止查看資源
  kubectl get pod -n kube-system
Error from server (Forbidden): pods is forbidden: User "pod-viewer" cannot list resource "pods" in API group "" in the namespace "kube-system"
```

大功告成～！

## 結論

到此我們就已經順利的完成了一個完整的 `RBAC` 流程了，`Kubernetes` 中在講述權限的篇幅其實非常多，不只是需要對 `Kubernetes` 資源有一定的了解，更要對認證與授權這些大觀念有深入的概念才可以一窺其妙，加上各種主流或少見的驗證方法，是相對進階的課題了，有興趣的同學真的很推薦把關於這邊的官方文件都啃過一遍，肯定可以跟我一樣讀的頭破血流的^_^。

相關程式碼同時收錄在：

[https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day29](https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day29)

Reference

****[管理服务账号](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/service-accounts-admin/)****

[https://github.com/kvaps/kubectl-node-shell](https://github.com/kvaps/kubectl-node-shell)

****[[Day17] k8s管理篇（三）：User Management、RBAC、Node Ｍaintenance](https://ithelp.ithome.com.tw/articles/10223717)****

****[基於角色的訪問控制（RBAC）](https://jimmysong.io/kubernetes-handbook/concepts/rbac.html)****

****[RBAC with Kubernetes in Minikube](https://medium.com/@HoussemDellai/rbac-with-kubernetes-in-minikube-4deed658ea7b)****

****[openSSL 自發憑證](https://hackmd.io/@yzai/rJXYxFpmq)****

****[【從題目中學習k8s】-【Day20】第十二題 - RBAC](https://ithelp.ithome.com.tw/articles/10244300)****

****[使用RBAC 鑑權](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/rbac/#referring-to-subjects)****

****[k8s 基於RBAC的認證、授權介紹和實踐](https://iter01.com/657295.html)****

****[Day 19 - 老闆！我可以做什麼：RBAC](https://ithelp.ithome.com.tw/articles/10195944)****

**[了解 Kubernetes 中的認證機制](https://godleon.github.io/blog/Kubernetes/k8s-API-Authentication/)**

****[用戶認證](https://kubernetes.io/zh-cn/docs/reference/access-authn-authz/authentication/)****