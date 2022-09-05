# Day5 Kubernetes Dashboard 你的 Kubernetes GUI 神器

## 概述

在各種與容器以及服務設定的 Kubernetes 世界裡，我們會非常非常頻繁的使用 `kubectl` 指令去對 `kube-apiserver` 去做請求並且回傳資訊到我們的終端機介面中，於是官方就推出了一套 `Kubernete Dashboard` 做為 Web UI 工具給我們，不只可以一目了然的列出所有容器的服務狀態，並且可以在上面可以將我們的 kubectl 指令轉變成在 UI 上的功能，可以說是在本地端使用 Kubernetes 的標配了(通常在雲端平台會搭配雲端自家的監控整合介面)。

![https://ithelp.ithome.com.tw/upload/images/20220905/201495629zt8FRdn7d.png](https://ithelp.ithome.com.tw/upload/images/20220905/201495629zt8FRdn7d.png)

### 配置 ****Kubernetes dashboard****

1. 下載 Kubernetes dashboard 相關容器服務

```bash
kubectl apply -f https://raw.githubusercontent.com/kubernetes/dashboard/v2.5.1/aio/deploy/recommended.yaml
```

1. 檢查 kubernetes-dashboard 應用狀態

```bash
kubectl get pod -n kubernetes-dashboard
```

![https://ithelp.ithome.com.tw/upload/images/20220905/20149562r75rtO5FlH.png](https://ithelp.ithome.com.tw/upload/images/20220905/20149562r75rtO5FlH.png)

1. 開啟 API Server 訪問代理並訪問 Kubernetes dashboard 頁面：

```bash
kubectl proxy
-------
Starting to serve on 127.0.0.1:8001
```

通過如下 URL 訪問 Kubernetes dashboard

[http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/](http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:kubernetes-dashboard:/proxy/)

這時畫面將會出現需要訪問權限 token 的驗證：

![https://ithelp.ithome.com.tw/upload/images/20220905/201495621ZZj5uaX4G.png](https://ithelp.ithome.com.tw/upload/images/20220905/201495621ZZj5uaX4G.png)

1. 接下來我們需要產生一個預設 Token 來登入頁面，直接在終端機輸入下列指令：

```bash
kubectl apply -f - <<EOF
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kube-system-default
  labels:
    k8s-app: kube-system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
  - kind: ServiceAccount
    name: default
    namespace: kube-system

---

apiVersion: v1
kind: Secret
metadata:
  name: default
  namespace: kube-system
  labels:
    k8s-app: kube-system
  annotations:
    kubernetes.io/service-account.name: default
type: kubernetes.io/service-account-token
EOF
```

1. 打印出 Token (MacOs)：

```bash
TOKEN=$(kubectl -n kube-system describe secret default| awk '$1=="token:"{print $2}')

kubectl config set-credentials docker-desktop --token="${TOKEN}"

echo $TOKEN
// your token...
```

1. 將上一步拿到的 Token 填入後，即可成功進入 Kubernetes Dashboard 主頁面！

![https://ithelp.ithome.com.tw/upload/images/20220905/201495621Bfpv7uhIV.png](https://ithelp.ithome.com.tw/upload/images/20220905/201495621Bfpv7uhIV.png)

## 結論

Kubernetes Dashboard 不得不說是一個初學者以及指令苦手的救星，在初期對 Kubernetes 完全不熟悉的我，甚至連常用的指令都不清楚。有了 Kubernetes Dashboard 這個 GUI 幫助，讓我從另一個角度去更了解到 Kubernetes 的運作方式，同時也讓我學到更多相關指令，當然有很多唾棄 GUI 的指令神人會覺得圖形化工具會寵壞工程師，但莫忘世上苦人多，更多的是對著較高的學習門檻打退堂鼓的平庸小白，所以我反而認為圖性化工具提供一個機會讓我們從另一個層面去學習一個工具，並且跟指令操作可以相輔相成。