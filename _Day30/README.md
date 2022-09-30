# 從異世界歸來的第三十天 - Kubernetes 最終章 - 學習了 Kubernetes 的下一步呢？

## 前言

不敢相信自己就這樣從第一天開賽撐到了第三十天。還記得今年的五月底，有幸被一位大神朋友邀請加入他的鐵人賽隊伍，雖然後來大神與他的快樂夥伴因公務繁忙沒辦法參加，於是後來就演變成我一人背負著全村的希望來挑戰個人組，簡稱被放生。

當時是我到了新工作中接觸 `Kubernetes` 不久，感覺有股神秘的力量默默的指引著我，我知道是它選擇了我，於是我埋首緊跟不敢有任何怠慢。這是一場長達四、五個月旅途，在漸漸互相熟悉的過程中，重新見到許多過去未曾熟絡的老戰友。

因為接觸了容器化技術遇到了 Docker，因為熟悉了 Docker 而勇敢跳進 `Kubernetes` 的坑，因為前端團隊處於混沌狀態而去學習了 React 來試圖將前端導回正軌， 因為熟悉了前端後端以及 `Kubernetes` 而自告奮勇的，為團隊從零到有建立了 Gitlab CI/CD 多環境部署，在了解 CI/CD 的過程領悟到了資料庫版本控制以及建立應用服務測試環境在 CI 中的重要性以及難點，後來又反思了單體式應用以及微服務應用在開發以及部署中的差異，直到我著手練習了第一個微服務練習，多線程高併發以及資料庫的最終一致性這些都是我半年前都是始料未及的，直到最近因為公司需要，獨自架了一個直播串流服務，了解了影音串流花俚胡哨的專有名詞和格式，被軟解轉碼所消耗的龐大效能打了一巴掌後，除了開始了解硬體效能以及 CPU 與 GPU 如何運作並各司其職，也開始注意到了資源監控以及 Logging 這個兩個大 Boss，還有自己的不足。

寫了三個月的文章終於在完賽的前兩天湊齊最後一篇，雖然說終於也是另一種開始，開始了追逐，開始無條件付出，也開始沒有退路，追的太快有時候太超過，突然之間回頭才發現初衷已飛走。

有時候發現長越大夢想卻越小，自己究竟幹了些什麼，又拿時間換了什麼，長越大才了解實現的定義是成為現實，有時候決定下得太慎重反而讓後悔成為偶爾的陣痛，於是我決定不再去想那麼多，從異世界回到了現實，只要一股腦的變強就好了吧？唯一不變的，是我仍然是當初那個油箱全滿的死文組。

## 從異世界歸來的下一步呢

即使我打從最一開始就知道 `Kubernetes` 是一條不歸路，死命活命腦霧學習個幾個月，僅僅也只是容器化世界的幾粒細沙而已，每次說到這裡我都會請 Kubernetes 群組裡的小夥伴們支援冰山一角圖 XD。

![https://ithelp.ithome.com.tw/upload/images/20220929/20149562yQWiFhU8xd.png](https://ithelp.ithome.com.tw/upload/images/20220929/20149562yQWiFhU8xd.png)

從中我們可以很清楚的得知，本系列文分享的進度大概只位在於冰山深淵的第二層跟第三層之間，如果你也一樣跟我在螢幕面前絕望過，請繼續保持良善繼續革命，繼續凝視著深淵，因為深淵同時也凝視著你，或許有天絕望了第十一次，也可以跟國父一樣收穫成功的果蕾。

接下來將跟各位分享一下自己準備繼續鑽研的路線，可以簡單的分為三個部分： `Templating` 、`Server Mesh` 、`Monitoring` ，這些可以說是得益於容器化以及微服務下的時代潮流。有了 `Kubernetes` 這種容器編排工具使得多環境部署更加方便，催生出了可以重複使用降低耦合度的 `Kubernetes` 設定檔模版 - `Helm` ，而微服務的興起也代表著每個服務之間的溝通過程越發重要，也使得有個統一溝通中心的 `Server Mesh` 的概念被提出，其中最具代表性的工具一定是 `istio` ，而當微服務越發龐大複雜時，排查錯誤的難度也指數上升，此時我們需要一個超脫一切生命週期的地方井然有序的儲存我們的日誌，並且時刻監控資源，這時許多人一定會想到 `Grafana` 和 `Prometheus` 這套威猛組合拳，光是以上提及的服務就足已使我們踏入第四與第五層深淵，更足以讓我們繼續螫疼下一個半年。

### Helm

![https://ithelp.ithome.com.tw/upload/images/20220929/20149562HgVrEQlEdO.png](https://ithelp.ithome.com.tw/upload/images/20220929/20149562HgVrEQlEdO.png)

Helm是一個模板引擎，允許您編寫可重用的Kubernetes模板。

假如我們的服務有分正式環境以及測試環境，其中的差別就只在於標籤的值，而 `Helm` 可以讓我們相關參數設定成變數，進而使用一個管理變數值的設定並重用所有在不同環境中沒有差異的部分。

實際查看例子：

```yaml
# 原版的正式環境設定檔，測試環境需要產生另一個相同檔案來設定 environment: development
apiVersion: v1
kind: Pod
metadata:
  name: busybox-sleep
  labels:
    environment: production
spec:
  containers:
    - name: busybox
      image: busybox
      args:
        - sleep
        - "1000000"
```

如果我們將 `environment` 的值設為變數：

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox-sleep
  labels:
    environment: {{ .Values.environment }}
spec:
  containers:
    - name: busybox
      image: busybox
      args:
        - sleep
        - "1000000"
```

即可將多個設定檔統一，並將變數值也統一管理。

### Istio

![https://ithelp.ithome.com.tw/upload/images/20220929/20149562w2ZiXj8QBg.png](https://ithelp.ithome.com.tw/upload/images/20220929/20149562w2ZiXj8QBg.png)

Istio 是一種服務網格，也就是一種現代化的服務網路層，提供透明化且與程式設計語言種類無關的平台，能讓您以有彈性又簡單的方式將應用程式網路功能自動化。

聽起來很抽象，可以簡單理解為，將服務之間的溝通扁平化而達到服務管理的一致性、更細膩的第七層流量管理、負載均衡、鏈路追蹤以及日誌收集和監控。

![https://ithelp.ithome.com.tw/upload/images/20220929/20149562v1pi0fTtFV.png](https://ithelp.ithome.com.tw/upload/images/20220929/20149562v1pi0fTtFV.png)

### ****Grafana & Prometheus****

![https://ithelp.ithome.com.tw/upload/images/20220929/20149562549wkBCNWz.png](https://ithelp.ithome.com.tw/upload/images/20220929/20149562549wkBCNWz.png)

得利於上面提到的 `Istio` 對服務溝通的強項，使得收集資源狀態和日誌紀錄的接口更加統一方便，`Istio` 甚至替我們預設好了 Grafana 和 Prometheus 插件，大大降低了我們進入 `Monitoring` 領域的門檻。

`Grafana & Prometheus` ********這兩套工具的組合，目前是市場上非常成熟主流的監控解決方案，而關於這個領域的水也非常的深，千言萬語就用下面一張圖去呈現出來，其他的就留給各位有機會的時候去細細品嚐了。

![https://ithelp.ithome.com.tw/upload/images/20220929/20149562WfbIkG0mog.png](https://ithelp.ithome.com.tw/upload/images/20220929/20149562WfbIkG0mog.png)

## 結語

恭喜同時在為鐵人賽奮鬥的勇士們，只有經歷過才知道到山頂之前的風景，也感謝此時在觀看文章的各位。前後三個多月的時間，每天強迫自己多看一點東西盡量在開賽前多完成一篇文章，看的愈多發現自己不懂得愈多，寫到後面甚至會開始感到疲乏，除了自己腦力不足很痛苦之外，更多的是會被路邊的鮮花轉移注意力，時間有限但慾望無窮，這段時間積了好多跟 `Kubernetes` 之外的東西想學但沒時間學啊啊啊。

總之，看來我還是撐過來了吧！一定給自己一個大拇指的啦，每天腦中的待辦事項暫時可以少了一個寫文章了灑花，剛好最近也算是一個踏進人生另一個階段時間點。在這裡預祝大家工作順利、寫文章順利、學習順利，如果有機會的話，搞不好明年還會看到小弟跑來自虐喔。