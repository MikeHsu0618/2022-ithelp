從異世界歸來的第十八天 - Kubernetes Volume (三) - ConfigMap
---

## 概述

在正式環境中的產品開發中，大多人都會遇到不同環境的佈署，最簡單的會被分成開發環境（Development）以及正式環境（Production)，這時區分這些環境的關鍵往往是連線的資料庫或是使用的 Token 、 ApiKey 或是初始化資料等等，這些被抽離出來的設定(Configuration)很好的降低了程式碼的耦合，讓我們只需要簡單的編輯一下相關設定檔即可創造出我們想要的相對環境，而 `Kubernetes` 提供了 `ConfigMap` 讓我們可以從最頂層方便的向下注入。

## ConfigMap 的特性

基本上 `ConfigMap` 可以理解為存放設定檔的功能，也就是說這個物件可以直接連接一個或多個檔案，在我們程式中需要的地方可以使其被取用。

- 一個 `ConfigMap` 物件可以存入一個或多個設定檔。
- 降低程式碼耦合，同一份代碼只需抽換不同的 `ConfigMap` 即可切換不同環境。
- 統一存放所有設定檔，在 `Kubernetes` 中可以統一查看以及管理。

## 建立 ConfigMap

簡單介紹建立一個 `ConfigMap` 的幾種方式：

### 1. 用指令匯入整個檔案 :

```jsx
# initdb.sql
DROP TABLE IF EXISTS posts CASCADE;

CREATE TABLE posts
(
    id             BIGSERIAL PRIMARY KEY,
    uuid           VARCHAR(36)  NOT NULL UNIQUE,
    user_id        NUMERIC      NOT NULL,
    title          VARCHAR(255) NOT NULL,
    content        TEXT         NOT NULL,
    comments_count NUMERIC               DEFAULT 0,
    created_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMP    NULL
);

CREATE INDEX user_id_key ON posts (user_id);

COMMENT ON COLUMN posts.title IS '標題';
    COMMENT ON COLUMN posts.content IS '內容';
    COMMENT ON COLUMN posts.comments_count IS '評論數';
```

```jsx
kubectl create configmap pg-initsql --from-file=initdb.sql
--------
configmap/pg-initsql createdg-initsql
```

利用 `kubectl create` 指令將整個檔案設定成一個 `configMap`

<aside>
? 看到這裡可能有些人對 `kubectl apply` 跟 `kubectl create` 兩者有點混淆，但可以簡單的理解成 `kubectl create` 明確的告訴 `Kubernetes` 他將建立一個資源物件，而 `kubectl apply` 則通常伴隨著 yaml 設定檔表示該物件應該要怎麼什麼樣子。

</aside>

查看一下產生結果：

```jsx
kubectl describe configmap pg-initdb
---------
Name:         pg-initdb
Namespace:    default
Labels:       <none>
Annotations:  <none>

Data
====
initdb.sql:
----
DROP TABLE IF EXISTS posts CASCADE;

CREATE TABLE posts
(
    id             BIGSERIAL PRIMARY KEY,
    uuid           VARCHAR(36)  NOT NULL UNIQUE,
    user_id        NUMERIC      NOT NULL,
    title          VARCHAR(255) NOT NULL,
    content        TEXT         NOT NULL,
    comments_count NUMERIC               DEFAULT 0,
    created_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMP    NULL
);

CREATE INDEX user_id_key ON posts (user_id);

COMMENT ON COLUMN posts.title IS '標題';
    COMMENT ON COLUMN posts.content IS '內容';
    COMMENT ON COLUMN posts.comments_count IS '評論數';

BinaryData
====
```

### 2. 使用指令建立 key-value 組合：

```jsx
kubectl create configmap pg-connect \
--from-literal=host=127.0.0.1 \
--from-literal=port=5432
```

以上我們使用指令建立 `host` `port` 兩組 key-value。

查看一下結果：

```jsx
kubectl describe configmap pg-connect
----------
Name:         pg-connect
Namespace:    default
Labels:       <none>
Annotations:  <none>

Data
====
host:
----
127.0.0.1
port:
----
5432

BinaryData
====
```

### 3. 使用 yaml 檔建立設定檔：

```jsx
# initdb-configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: initdb-yaml
  labels:
    app: db
data:
  initdb.sql: |
    DROP TABLE IF EXISTS posts CASCADE;
    
    CREATE TABLE posts
    (
        id             BIGSERIAL PRIMARY KEY,
        uuid           VARCHAR(36)  NOT NULL UNIQUE,
        user_id        NUMERIC      NOT NULL,
        title          VARCHAR(255) NOT NULL,
        content        TEXT         NOT NULL,
        comments_count NUMERIC               DEFAULT 0,
        created_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
        updated_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
        deleted_at     TIMESTAMP    NULL
    );
    
    CREATE INDEX user_id_key ON posts (user_id);
    
    COMMENT ON COLUMN posts.title IS '標題';
    COMMENT ON COLUMN posts.content IS '內容';
    COMMENT ON COLUMN posts.comments_count IS '評論數';
```

```jsx
kubectl apply -f initdb-configmap.yaml
---------
configmap/post-initdb-yaml created
```

查看一下結果：

```jsx
kubectl describe configmap initdb-yaml
---------
Name:         initdb-yaml
Namespace:    default
Labels:       app=db
Annotations:  <none>

Data
====
initdb.sql:
----
DROP TABLE IF EXISTS posts CASCADE;

CREATE TABLE posts
(
    id             BIGSERIAL PRIMARY KEY,
    uuid           VARCHAR(36)  NOT NULL UNIQUE,
    user_id        NUMERIC      NOT NULL,
    title          VARCHAR(255) NOT NULL,
    content        TEXT         NOT NULL,
    comments_count NUMERIC               DEFAULT 0,
    created_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMP    NULL
);

CREATE INDEX user_id_key ON posts (user_id);

COMMENT ON COLUMN posts.title IS '標題';
COMMENT ON COLUMN posts.content IS '內容';
COMMENT ON COLUMN posts.comments_count IS '評論數';

BinaryData
====
```

### 4. 使用 yaml 檔建立 key-value：

```jsx
# initdb-kv.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: initdb-kv-yaml
  labels:
    app: db
data:
  PG_USER: postgres
  PG_PASSWORD: postgres
```

查看一下結果：

```jsx
kubectl describe configmap initdb-kv-yaml
--------
Name:         initdb-kv-yaml initdb-kv-yaml          
Namespace:    default
Labels:       app=db
Annotations:  <none>

Data
====
PG_PASSWORD:
----
postgres
PG_USER:
----
postgres

BinaryData
====
```

## 實際應用 Config

```jsx
# pg-pod.yaml
apiVersion: v1
kind: Pod
metadata:
  name: db
  labels:
    app: db
spec:
  containers:
    - name: db
      image: postgres:12.4-alpine
      env:
        # 使用 configmap 的 key-value 做為值傳入
        - name: POSTGRES_USER
          valueFrom:
            configMapKeyRef:
              name: initdb-kv-yaml
              key: PG_USER
        - name: POSTGRES_PASSWORD
          valueFrom:
            configMapKeyRef:
              name: initdb-kv-yaml
              key: PG_PASSWORD
        - name: PGDATA
          value: '/var/lib/postgresql/data/pgdata'
        - name: POSTGRES_DB
          value: 'posts'
      ports:
        - containerPort: 5432
      volumeMounts:
        # 使用 configmap 做為 file 當作初始化設定
        - mountPath: /docker-entrypoint-initdb.d
          name: initdb
  volumes:
    - name: initdb
      configMap:
        name: initdb
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: initdb
  labels:
    app: db
data:
  initdb.sql: |
    DROP TABLE IF EXISTS posts CASCADE;
    
    CREATE TABLE posts
    (
        id             BIGSERIAL PRIMARY KEY,
        uuid           VARCHAR(36)  NOT NULL UNIQUE,
        user_id        NUMERIC      NOT NULL,
        title          VARCHAR(255) NOT NULL,
        content        TEXT         NOT NULL,
        comments_count NUMERIC               DEFAULT 0,
        created_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
        updated_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
        deleted_at     TIMESTAMP    NULL
    );
    
    CREATE INDEX user_id_key ON posts (user_id);
    
    COMMENT ON COLUMN posts.title IS '標題';
    COMMENT ON COLUMN posts.content IS '內容';
    COMMENT ON COLUMN posts.comments_count IS '評論數';
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: initdb-kv-yaml
  labels:
    app: db
data:
  PG_USER: PG_USER
  PG_PASSWORD: PG_USER
```

```jsx
kubectl apply -f pg-pod.yaml
-----------
pod/db created
configmap/initdb unchanged
configmap/initdb-kv-yaml unchanged
```

在上面的設定中我們同時利用的 `ConfigMap` 的檔案掛載以及 key-value 掛載，設定了我們的 `postgres` 初始化資料表以及使用者帳號密碼。

接下來我們來驗證一下是否設定都已成功注入到 `postgres` 中：

### 進入容器：

```jsx
kubectl exec -it db -- sh

ls
---------
bin                         etc                         media                       proc                        sbin                        tmp
dev                         home                        mnt                         root                        srv                         usr
docker-entrypoint-initdb.d  lib                         opt                         run                         sys                         var
```

成功進入 `postgres` 即可使用 psq cli 操作。

### 檢查 initdb.sql 是否成功掛載：

```jsx
cat docker-entrypoint-initdb.d/initdb.sql
---------
DROP TABLE IF EXISTS posts CASCADE;

CREATE TABLE posts
(
    id             BIGSERIAL PRIMARY KEY,
    uuid           VARCHAR(36)  NOT NULL UNIQUE,
    user_id        NUMERIC      NOT NULL,
    title          VARCHAR(255) NOT NULL,
    content        TEXT         NOT NULL,
    comments_count NUMERIC               DEFAULT 0,
    created_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMP    NOT NULL DEFAULT NOW(),
    deleted_at     TIMESTAMP    NULL
);

CREATE INDEX user_id_key ON posts (user_id);

COMMENT ON COLUMN posts.title IS '標題';
COMMENT ON COLUMN posts.content IS '內容';
COMMENT ON COLUMN posts.comments_count IS '評論數';
```

成功獲得 `ConfigMap` 檔案。

### 檢查 Role 是否被成功創立：

在剛剛的 `postgres` 容器中，使用 psql cli 登入：

```jsx
psql -U PG_USER -d posts
-------
psql (12.4)
Type "help" for help.

posts=#
```

成功以我們設定的帳號名稱 `PG_USER` 以及資料庫 `posts` 登入。

### 檢查是否成功建立資料表 Posts

```jsx
posts=# \d posts
                                          Table "public.posts"
     Column     |            Type             | Collation | Nullable |              Default              
----------------+-----------------------------+-----------+----------+-----------------------------------
 id             | bigint                      |           | not null | nextval('posts_id_seq'::regclass)
 uuid           | character varying(36)       |           | not null | 
 user_id        | numeric                     |           | not null | 
 title          | character varying(255)      |           | not null | 
 content        | text                        |           | not null | 
 comments_count | numeric                     |           |          | 0
 created_at     | timestamp without time zone |           | not null | now()
 updated_at     | timestamp without time zone |           | not null | now()
 deleted_at     | timestamp without time zone |           |          | 
Indexes:
    "posts_pkey" PRIMARY KEY, btree (id)
    "posts_uuid_key" UNIQUE CONSTRAINT, btree (uuid)
    "user_id_key" btree (user_id)
```

大功告成～

## 結論

今天我們實作了 `ConfigMap` 最常見的兩種使用方式，在官方文件中 `ConfigMap` 還有其他更進階的操作方式，這裡就先簡單介紹可以應付多數應用場景的方法。

相關文章：

- [從異世界歸來的第十六天 - Kubernetes Volume (一) - Volume 是什麼](https://ithelp.ithome.com.tw/articles/10291557)

相關程式碼同時收錄在：

[https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day18](https://github.com/MikeHsu0618/2022-ithelp/tree/master/Day18)

Reference

****[配置Pod 使用ConfigMap](https://kubernetes.io/zh-cn/docs/tasks/configure-pod-container/configure-pod-configmap/)****

****[[Day 18] 高彈性部署 Application - ConfigMap](https://ithelp.ithome.com.tw/articles/10196153)****

****[Kubernetes 那些事 — ConfigMap 與 Secrets](https://medium.com/andy-blog/kubernetes-%E9%82%A3%E4%BA%9B%E4%BA%8B-configmap-%E8%88%87-secrets-5100606dd06c)****