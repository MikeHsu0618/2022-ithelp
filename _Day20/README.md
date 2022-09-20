從異世界歸來的第二十天 - Kubernetes Volumes (五) - PV & PVC
---

## 概述

`PersistentVolume` 與 `PersitentVolumeClaim` (以下簡稱 PV ＆ PVC)的觀念通常很容易的與如何使服務 `Stateful` 的設定掛上關係，最重要的關鍵是 `PV` 的生命週期獨立於 `Pod` 之外，使得我們在消滅或擴展 `Pod` 時可以繼續持有原有的資料，而這些資料儲存的地方也從原本的 `Pod` 中，轉而由 NFS(Network File System)、叢集中的 `Node` 、雲端儲存服務等相關儲存空間來進行存放，進而實現獨立的生命週期。

### ****Persistent Volumes (PV)****

`Kubernetes` 利用 `PV` 提供一個抽象的存儲空間，並且 `PV` 能被動態和靜態的被提供，可以簡單理解為當 `PV` 是預先被宣告出來後被 `PVC` 取用的話就是一種靜態，而如果 `PVC` 中有指定 `storageclass` 的種類時， `Kubernetes` 將會動態的為我們產生 `PV`。

當我們作為一位系統管理者宣告了一塊存儲空間後，而系統的使用者就能以 `PVC` 來請求此存儲空間，形成下圖的關係：



### ****Persistent Volume Claims (PVC)****

在 `PVC` 中表達的是使用者對存儲的請求，相較於 `Pod` 可以請求特定數量的 CPU 和內存資源， `PVC` 也可以請求存儲空間大小以及設定訪問模式 (例如：ReadWriteOnce、ReadOnlyMany 或 ReadWriteMany)，而 `PVC` 的資源請求成立後，將會去不斷的找尋符合條件的 `PV` 直到找到符合條件的資源並且將兩者綁定，如果找不到匹配的 `PV` 時， `PVC` 將會無限期的處於未榜定狀態(Pending)，直到出現與之匹配的 `PV` 加入。

### 實際操作

```jsx
# pvc.yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: pvc-demo
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: hostpath
  resources:
    requests:
      storage: 1Gi
```

以上我們使用 `PVC` 動態產生 `PV` 並且綁定兩者。

訪問模式(spec.accessModes)：

- **`ReadWriteOnce`**卷可以被一個節點以讀寫方式掛載。ReadWriteOnce 訪問模式也允許運行在同一節點上的多個Pod 訪問卷。
- **`ReadOnlyMany`**卷可以被多個節點以只讀方式掛載。
- **`ReadWriteMany`**卷可以被多個節點以讀寫方式掛載。
- **`ReadWriteOncePod`**卷可以被單個Pod 以讀寫方式掛載。如果你想確保整個集群中只有一個Pod 可以讀取或寫入該PVC， 請使用ReadWriteOncePod 訪問模式。

訪問模式是 `PV ＆ PVC` 中蠻值得注意的點，因為本系列文主要都是使用『單節點』的 Kubernetes (docker-desktop)，並不會遇到不同節點中的 `Pod` 掛載 `PV` 的情況，而現實的生產環境裡，往往的我們使用的 `Google GKE` 或 `AWS ELK` 通常都是多節點的情況，這時不同節點之間 `Pod` 以及服務就必須指定 **`ReadOnlyMany`** 或 **`ReadWriteMany`**  才能順利取得共享資源，加上支持 **`ReadOnlyMany` `ReadWriteMany`** 的 Provisioner 選擇不多，其中多半是使用雲端的 NFS 服務實現。

<aside>
💡 即使是相同服務的 Pod，也不一定會 `Kubernetes` 的資源分配器分配到同一個節點上，除非使用 `nodeSelector` 等顯式隱性設定。

</aside>

存儲類型(spec.storageClassName)：

在 `docker-desktop` 中預設的 `storageClass` 為 `hostpath` ，可以很直觀的看出這個 `storageClass` 很大的概率是將其的存儲空間設置在節點上，所以我們在 `docker-desktop` 的這個單節點裡可以實現出 `Pods` 之間的共享數據 。 如果管理員所創建的所有靜態 `PV` 都無法與用戶的 `PVC` 匹配， 集群可以嘗試為該 `PVC` 的設置動態製備一個存儲卷。這一製備操作是基於 `StorageClass` 來實現的，如 `PVC` 沒有特別指定則會使用預設 `StorageClass` 。 `PVC` 必須請求某個 `StorageClass` ， 同時集群管理員必須已經創建並配置了該類，這樣動態製備卷的動作才會發生。如果PVC 申領指定存儲類為`""`，則相當於為自身禁止使用動態製備的捲。

接下來產生我們的 `PVC` ：

```jsx
kubectl apply -f ./pvc.yaml
----------
persistentvolumeclaim/pvc-demo created
```

查看結果：

```jsx
kubectl get pv
----------

NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM              STORAGECLASS   REASON   AGE
pvc-c197c285-d314-43db-8cbc-6f912d8a9680   1Gi        RWO            Delete           Bound    default/pvc-demo   hostpath                2s

kubectl get pvc
----------
NAME       STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS   AGE
pvc-demo   Bound    pvc-c197c285-d314-43db-8cbc-6f912d8a9680   1Gi        RWO            hostpath       11s
```

可以看到 `PV ＆ PVC` 都已經成功的綁定(Bound) 並且配置相關的資源。

這裡 `PV` 會有四種狀態 (STATUS)

1. `Available`：表示 PV 為可用狀態
2. `Bound`：表示已綁定到 PVC
3. `Released`：PVC 已被刪除，但是尚未回收
4. `Failed`：回收失敗

`PV` 有三種回收策略 (RECLAIM POLICY)，分別是

1. `Retain`：手動回收
2. `Recycle` (已棄用)：透過刪除命令 `rm -rf /thevolume/*` ，取而代之的建議方案是動態產生。
3. `Delete`：刪除 PV 的同時也會一併刪除後端儲存磁碟。

這時我們再度拿出前幾天使用的 `emptyDir volume` 範例，將他由同個 `Pod` 分離成不同 `Pod` 來查看 `PV ＆ PVC` 的生命週期是否獨立於 `Pod` 。

```jsx
# nginx-pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
spec:
  containers:
    - name: nginx
      image: nginx:latest
			volumeMounts:
				- name: html
          mountPath: /usr/share/nginx/html
    volumes:
      - name: html
        persistentVolumeClaim:
          claimName: pvc-demo
          readOnly: false
```

```jsx
# alpine-pod.ymal
apiVersion: v1
kind: Pod
metadata:
  name: alpine-pod
spec:
  containers:
    - name: alpine
      image: alpine
      command: [ "/bin/sh", "-c" ]
      args: # 每十秒定時向 /html/index.html 寫入資料
        - while true; do
          echo $(hostname) $(date) >> /html/index.html;
          sleep 10;
          done
			volumeMounts:
        - name: html
          mountPath: /html
    volumes:
      - name: html
        persistentVolumeClaim:
          claimName: pvc-demo
          readOnly: false
```

將以上兩個掛在了 `PVC` 的 `Pod` 運行起來：

```jsx
kubectl apply -f ./nginx-pod.yaml ./alpine-pod.yaml
-----------
pod/alpine-pod configured
pod/nginx-pod configured
```

### 訪問 Pod 中的 Nginx

來確認一下 `alpine` 容器每隔 10 秒像 html/index.html 寫入訊息，而 `Nginx` 容器掛載的 `PVC` 是否同時也可以取得更新。

將 port 導出到本地的 [localhost](http://localhost) :

```jsx
kubectl port-forward pod/nginx-pod 8080:80
-------
Forwarding from 127.0.0.1:8080 -> 80pod 8080:80
Forwarding from [::1]:8080 -> 80
```

使用 curl 查看返回值：

```jsx
curl http://localhost:8080
--------
alpine-pod Sat Aug 6 03:31:50 UTC 2022
alpine-pod Sat Aug 6 03:32:00 UTC 2022
alpine-pod Sat Aug 6 03:32:10 UTC 2022
alpine-pod Sat Aug 6 03:32:20 UTC 2022
alpine-pod Sat Aug 6 03:32:30 UTC 2022
```

順利取得由 `alpine` 容器產生的內容～

## 結論

`PV & PVC` 的觀念很廣牽涉的範圍也很宏觀，理解起來需要對 `Kubernetes` 以及服務的運作場景都要有不少的理解。這幾天認真的鑽研了 `PV & PVC` 的觀念後，剛好公司就需要用到。只能說保持資料的持久性這方面真的是一塊大議題，在 StackOverflow 中有網友分享道： `Kubernetes` 的精神更適合拿用來做 `Stateless` 的微服務，並且將需要持久性的資料抽象出來做為一個獨立服務並開放 API 以供其他服務取得。

在最近的工作上也有更深的一層的認同，在小弟為了讓 `Google GKE` 上的多節點叢集實現共享持久性數據，就讓我把 Google 提供的 Kubernetes Provision storage 服務都實作過一次，分別是建立在 `Google FileStore` 的雲端 NFS 服務(實現了 ReadWriteMany)，以及建立在 `Google Compute Engine Disk` 的雲端硬碟服務(實現了 ReadWriteOnce)，兩種服務也對應了不同的情況，日後有機會可以再做一篇分享。

相關文章：

- [從異世界歸來的第十六天 - Kubernetes Volume (一) - Volume 是什麼](https://ithelp.ithome.com.tw/articles/10291557)

相關程式碼同時收錄在：

[https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day20](https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day20)

Reference
**[Using the Compute Engine persistent disk CSI Driver](https://cloud.google.com/kubernetes-engine/docs/how-to/persistent-volumes/gce-pd-csi-driver)**

****[持久卷](https://kubernetes.io/zh-cn/docs/concepts/storage/persistent-volumes/#access-modes)****

****[Day 15 - 別再遺失資料了：Volume (2)](https://ithelp.ithome.com.tw/articles/10193550)****