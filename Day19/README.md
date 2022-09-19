# 從異世界歸來的第十九天 - Kubernetes Volume (四) - Secret

## 概述

在上一篇文章中我們提到了 `ConfigMap` 這個 `Kubernetes` 讓我們解耦程式碼複雜度以及統一管理設定檔的好工具，但由於 `ConfigMap` 是使用明碼儲存一些不敏感的資料，那我們的 `API Key` 和 `金鑰` 等敏感資料就不太適合了呢。於是 `Kubernetes` 提供我們另一種選擇 - `Secret` 類似於 `ConfigMap` 的使用方式但他能對敏感資料有多一層保護而不會隨便暴露，但有趣的是 `Secret` 某方面來說也不是像是字面上的那麼安全，下面我們將會簡單談到他的運作原理。

## ****什麼是 Secret ?****

`Secret` 與 `ConfigMap` 在基本操作上大致相同，但在使用的方面不太一樣，所以這裡來特別提一下兩者比較不同的地方。`Secret` 是 `Kubernetes` 提供開發者存放敏感資料的方式， `Kubernetes` 本身也使用了相同的機制來存放 `Access Token` ，並限制 API 的存取權限，確保不會有外部服務隨意操作。

`Secret` 大致有三種類型：

1. `Service Account`：由 k8s 自動建立並掛載到 Pod，用來存取 `Kubernetes` API 使用，你可以在 `/run/secret/kubernetes.io/serviceaccount` 目錄中找到。
2. `Opaque`：以 base64 編碼的 Secret，用來儲存 密碼、金鑰等等。
3. `docker-registry`：如果映像檔是放在私有的 Registry，就需要使用這種類型的 `Secret` 。

在 `Kubernetes` 存取資料有以下幾種常見的方式：

1. 將 `Secret` 當作環境變數使用。
2. 將 `Secret File` 掛載在 `Pod` 中的某個路徑下面使用。
3. 在 `Pod` 中加入 `docker-registry secret` 讓我們不用在拉取私人庫時都需要先 `docker login` ，簡單來說是儲存 `docker login` 的帳號密碼讓 `Kubernetes` 可以自動登入順利運作。

## 建立 Secret

在建立 `Secret` 的值時，我們都要必須先使用 `base64` 進行轉碼，而 `Kubernetes` 在我們正確掛載後會自動幫我們解碼回原本的值。

1. 將 `Secret` 轉換為 `base64` ：

   首先我們可以使用內建語法取得經過 `base64` 的字串

    ```jsx
    echo -n 'my-account' | base64
    echo -n 'my-password' | base64
    ----------
    bXktYWNjb3VudA==
    bXktcGFzc3dvcmQ=
    ```

   於是我們獲得了以上的 `bXktYWNjb3VudA==` `bXktcGFzc3dvcmQ=` 做為加密過的字串。

2. 創建 yaml 設定檔：

    ```jsx
    # secret.yaml
    apiVersion: v1
    kind: Secret
    metadata:
      name: test-secret
    data:
      username: bXktYWNjb3VudA==
      password: bXktcGFzc3dvcmQ=
    ```

    ```jsx
    kubectl apply -f secret.yaml
    ----------
    secret/test-secret created
    ```

3. 查看一下結果：

    ```jsx
    kubectl get secret test-secret
    ----------
    NAME          TYPE     DATA   AGE
    test-secret   Opaque   2      15s
    ```

    ```jsx
    kubectl describe secret test-secret
    ----------
    Name:         test-secret test-secret
    Namespace:    default
    Labels:       <none>
    Annotations:  <none>
    
    Type:  Opaque
    
    Data
    ====
    password:  11 bytes
    username:  10 bytes
    ```

   可以在 Data 內看到 `Secret` 的 key 值和 `base64` 加密後的大小。

4. 或者你想只使用 `kubectl` 創建：

    ```jsx
    kubectl create secret generic test-secret --from-literal='username=my-account' --from-literal='password=my-password'
    ```

   這裡一樣使用了 create 當作創建資源的聲明指令，並且還需要加上 `generic` 這個 `subcommand`   表示我們要使用本地資源或者是 key-value 創建 `Secret`。


### 實際應用 Secret

讓我們利用上面建立的 `Secret` 來實際測試：

```jsx
# secret-test-pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: secret-test-pod
spec:
  containers:
    - name: test-container
      image: nginx
      volumeMounts:
        - name: secret-volume
          mountPath: /etc/secret-volume
  volumes:
    - name: secret-volume
      secret:
        secretName: test-secret
```

1. 創建 Pod 並查看：

    ```jsx
    kubectl apply -f secret-test-pod.yaml
    -----------
    pod/secret-test-pod createde
    ```

    ```jsx
    kubectl get pod secret-test-pod
    ------------
    NAME              READY   STATUS    RESTARTS   AGE
    secret-test-pod   1/1     Running   0          57s
    ```

2. 進入 Pod 中查看：

    ```jsx
    kubectl exec -it secret-test-pod -- sh
    ```

   在容器中打印出 `Secret` 內容

    ```jsx
    ls /etc/secret-volume
    ------------
    password  username
    ```

    ```jsx
    cat /etc/secret-volume/username
    ------------
    my-account
    
    cat /etc/secret-volume/password
    ------------
    my-password
    ```


成功打印出 `base64` 解碼後的數值～

## 聊一聊關於 Secret 並不看起來那麼安全的這件事

經過上面的介紹後，可能有人已經可以察覺到，我們可以非常容易的看到 `Secret` 的原碼，只要有相關的權限即可，雖然他的內容經過了 `base64` 編碼，但基本上等同於明文。

所以說， `Kubernetes` 原生的 `Secret` 是非常簡單的，不是特別適合在大公司直接使用，對 `RBAC(角色權限)` 的挑戰也比較大。

針對上面提到的問題，其大概的解決方案不難想到如下幾種，etcd加密、API Server 嚴格權限限制以及強化 Node 權限管理以及系統安全，並且以上的方案都需要缺一不可，如此看來，我們需要付出非常繁重的成本才能讓原生 `Secret` 在嚴謹條件下達到保障。

所以社群以及雲端服務商都有提供一些解套方案讓我們可以參考一下：

- AWS Key Management Service
- Google Cloud KMS

因為沒有實際操作過，這裡就先點到為止，主要目的是可以讓大家從另一個角度去思考一個工具的利弊以及取捨，接下來我們還會繼續介紹其他常用的 `Volume` 類別，敬請期待～

Reference

****[使用Secret 安全地分發憑證](https://kubernetes.io/zh-cn/docs/tasks/inject-data-application/distribute-credentials-secure/)****

****[[Day 12] 敏感的資料怎麼存在k8s?! - Secrets](https://ithelp.ithome.com.tw/articles/10195094)****

****[Kubernetes 那些事 — ConfigMap 與 Secrets](https://medium.com/andy-blog/kubernetes-%E9%82%A3%E4%BA%9B%E4%BA%8B-configmap-%E8%88%87-secrets-5100606dd06c)****

****[Day 17 - 藏好你的秘密：Secret](https://ithelp.ithome.com.tw/articles/10193940)****

****[关于 K8s 的 Secret 并不安全这件事](http://liubin.org/blog/2021/04/14/k8s-secrets-secret/)****