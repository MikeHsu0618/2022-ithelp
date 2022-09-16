從異世界歸來的第十六天 - Kubernetes Volume (一) - Volume 是什麼
---

## 概述
相信有使用過 `Docker` 的同學們對 `Volume` 都不會太陌生，其功能簡單來說是用來保存容器內的資料，此路徑資料夾內容將會與容器外的指定資料夾產生連接，即意味著這兩個資源是互通的，此後只要容器內的資料夾做任何存取，容器外的指定資料夾的內容也會跟著改變，讓我們可以再多個容器間共享一樣的資源，並且非常重要的是，當容器被刪除時連結的資料夾以及裡面的檔案並不會被刪除，透過這個特性我們便能做到『刪除容器卻保留資料』的作用，以實現當容器重啟後可以迅速還原資料以及狀態的功能。

### 那 Kubernetes 的 Volume 是什麼？

![https://ithelp.ithome.com.tw/upload/images/20220916/20149562MwLEil5qro.png](https://ithelp.ithome.com.tw/upload/images/20220916/20149562MwLEil5qro.png)

會有如此標題是因為相較於 `Docker` ， `Kubernetes` 的 `Volume` 擁有更豐富的類型以及更嚴謹的管理概念，並且 `Kubernetes` 在 `Volume` 中加入了生命週期的概念，使得我們擁有與 `Pod` 共生共滅的 `臨時卷(Ephemeral Volumes)` 和生命週期比 `Pod` 還長的持久卷，使得我們在重啟容器期間數據都不會丟失。

`Volume` 的核心就是一個目錄資料夾，其中可能存有數據，Pod 中的容器可以訪問該目錄中的數據。所採用的特定的捲類型將決定該目錄如何形成的、使用何種介質保存數據以及目錄中存放的內容。

使用卷時, 在 `.spec.volumes` 字段中設置為Pod 提供的捲，並在 `.spec.containers[*].volumeMounts` 字段中聲明卷在容器中的掛載位置。

### ****Volume 類型****

由於 `Kubernetes` 官方提供的類型實在過於精細並且迭代非常快速（每隔幾版就被棄用了），許多東西碰到的機會可以說是非常少，以下就大致講一下幾種最常見到的 `Volume` 類型～

- **EmptyDir**：

  當新增一個 `Pod` 時， `Kubernetes` 就會在新增一個 `emptyDir` 的空白資料夾，讓此 `Pod` 中的所有容器都可以讀取這個特定的資料夾，常用於 `數據緩存` 以及 `臨時存儲` 。

  不過基於 `emptyDir` 建構的 `gitRepo Volume` 可以在 `Pod` 起始生命週期時從對應的 GitRepo 中複製相對應文件資料到底層的 `emptyDir` 中，這使得他也可以具有一定意義上的持久性。

- **HostPath**：

  `HostPath` 能將節點(Node)上的文件或目錄掛載到你的 `Pod` 中，雖然這種需求不會太常遇到，但是他為了一些應用提供了更強大的後勤功能，例如在運行一個指定的 `Pod` 之前，先確認某 `HostPath` 下的文件是否存在以及應該以什麼方式存在。

>     ? 目前官方指明 `HostPath` 存在許多安全風險，最佳做法是盡可能避免使用HostPath。當必須使用 `HostPath`  時，它的範圍應僅限於所需的文件或目錄，並以`只讀`方式掛載。

- **Network FileSystem(NFS)**：

  `NFS` 能將網路文件掛載在你的 `Pod` 中。不像 `empty` 那樣會在刪除 `Pod` 的同時被刪除，這意外著他可以當作預先填充的數據，並且使這些數據在 `Pod` 之間共享。通常會搭配雲端儲存空間的服務使用。

- **ConfigMap**：

  看到 `ConfigMap` 應該一看就知道這個物件跟某些設定有關，沒錯 `ConfigMap` 通常都是用來存放設定檔用的，也就是說這個物件當作我們常用的環境變數檔或者是資料庫初始化設定等等偏向佈署方面的用途。

- **Secrets**：

  而 `Secrets` 字面上就可以明白相較於 `ConfigMap` 用來存放偏向佈署面的檔案， `Serects` 通常是用來存放敏感的資料，像是使用者帳密、憑證等等，而他一樣也具備 `ConfigMap` 擁有的功能並且還有專屬於 `Serects` 才有的特性，如 Secret 會將內部資料進行 `base64` 編碼。
- **PV & PVC (PersistentVolume, PersistentVolumeClaim)**：

  `PV` 是集群中的一塊儲存資源，可以由管理者事先設定，是屬於集群的資源，所以他們擁有獨立於任何使用 `PV` 的 `Pod` 的生命週期。

  而 `PVC` 表達的是使用者對於儲存的請求，像是 `Pod` 會對 `Node` 的資源請求（CPU 或內存）， `PVC` 同樣的也會請求並且消耗 `PV` 的限額。


## 結論

在 `Kubernets` 中，如何資料的儲存與設置永遠都是一大課題跟學問，從我們都稍微比較熟悉的 `Volume` 個人覺得是一個很好的切入點，接下來的幾天我們將會開始實際作些操作練習，期望逐漸加深我們對於 `Volume` 的概念以及定位。

相關程式碼同時收錄在：

[https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day16](https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day16)

Reference

[Kubernetes 文檔](https://kubernetes.io/zh-cn/docs/)

****[Kubernetes中的emptyDir存儲捲和節點存儲卷](https://cloud.tencent.com/developer/article/1660415)****

****[Kubernetes 那些事 — ConfigMap 與 Secrets](https://medium.com/andy-blog/kubernetes-%E9%82%A3%E4%BA%9B%E4%BA%8B-configmap-%E8%88%87-secrets-5100606dd06c)****