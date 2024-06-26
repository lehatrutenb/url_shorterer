# url_shorterer

**Api**
----
  Returns short url for given long one

* **URL**

  :9090/post

* **Method:**

  `POST`

* **Data Params**

  `your_long_url: string`

* **Success Response:**

  * **Code:** 200 <br />
    **Body:** `short_url: string`
 
* **Error Response:**

  * **Code:** 400 StatusBadRequest <br />
   **Possible reason:** `incorrect body`

  OR

  * **Code:** 500 StatusInternalServerError <br />
    **Possible reason:** `server has no more short urls temporally`

* **Sample Call:**

  ```
  curl --location 'http://127.0.0.1:9090/post' \
  --header 'Content-Type: text/plain' \
  --data 'http://gmail.com'
  ```
  
----
  Returns short url for given long one

* **URL**

  :9091/*

* **Method:**

  `GET`

* **Required param:**
 
   `short_url=[string]`


* **Success Response:**

  * **Code:** 301 StatusMovedPermanently <br />
    **Location:** `long_url: string`
 
* **Error Response:**

  * **Code:** 404 StatusNotFound <br />
   **Possible reason:** `incorrect or not found url`

  OR

  * **Code:** 503 StatusServiceUnavailable <br />
    **Possible reason:** `server is too busy`

  OR

  * **Code:** 500 StatusInternalServerError <br />
    **Possible reason:** `smth went wrong`


* **Sample Call:**

  ```
  curl --location 'http://127.0.0.1:9091/4xabacaba'
  ```

# How to use it by yourself?
```
$ docker compose build ; to build images
$ docker compose -p url_shorterer up ; to run docker compose - may be slow as it use kafka
$ docker compose -p url_shorterer down ; to stop all containers
```

# How it works?

**Sequence diargams**
* ***Add url***
![telegram-cloud-document-2-5469930164748045806](https://github.com/lehatrutenb/url_shorterer/assets/36619154/f6adf6fd-60e2-4d2e-b860-4ffc31b9b705)
----
* ***Add url inside***
![telegram-cloud-document-2-5469930164748045740](https://github.com/lehatrutenb/url_shorterer/assets/36619154/19fa904f-0a38-4e5d-b692-d66d98da3e63)
----
* ***Get url***
![telegram-cloud-document-2-5469930164748045784](https://github.com/lehatrutenb/url_shorterer/assets/36619154/964ff7a3-16e9-494d-bffb-837f0f330a53)
----


