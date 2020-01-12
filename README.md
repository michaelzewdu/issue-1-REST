# Issue #1

A platform for fledgling authors and artists to publish their works, focusing on web-serials (web-novels) and web comics and allow them to directly interact with their audience.  

#### Guide (For Delivery Date Jan 12)

### https://github.com/slim-crown/issue-1-website
That links lead to another repository, a part of this project.
It's a website that consumes the REST API in this repo to function.

#### State of the project
This repo is currently build passing and almost all the routes are working. 
This includes the Authentication routes. The website repo on the other hand 
only has a half the consumers written for it and much work to be done on it.

##### Dependencies
We, unfortunately, found out about the dependencies restrictions
too late and it's with apologies that we present the following list. 
We've tried to prune the all non-fatal packages and the following are used
because they were found to be indispensable.

```shell script
go get github.com/gorilla/mux
go get github.com/satori/go.uuid
go get github.com/dgrijalva/jwt-go

# following are used for data sanitation
go get github.com/microcosm-cc/bluemonday
go get gopkg.in/russross/blackfriday.v2
```

##### Database
Due to somewhat complex database setup, we've prepared scripts to assist with setting it up
and they can be found in the **_setup** directory. The script, **script.bat**, is only a simple 
batch script and one can inspect it to setup the database without using the script.

##### Running

**main.go** can be found inside the **cmd\server** path but 
the project will have to be built from the root path for relative
routes to function.

Right where the server is started in main.go can be found a commented 
out line that can be used to start the server in HTTPS secured mode.

```go
//log.Fatal(http.ListenAndServeTLS(":"+setup.Port, "cmd/server/cert.pem", "cmd/server/key.pem",mux))
```

##### Routes

A complete list of the routes can be discerned by looking at the **pkg\delivery\http\rest\mux.go**.

###### Queries

Routes that `GET` a list of entities can usually be queried in the following forms:

```http request
GET http://localhost:8080/users?sort=last_name_asc&limit=5&offset=0&pattern=rem
```

- Sort queries are different from entity to entity and can be discerned by looking
at their respective service. `_ASC` or `_DSC` can be appended to customize the order.

- Limit & Offset can be used to specify precise pagination.

- Pattern can be used to search.
    - `/releases` and `/posts` support full text searching.
    - `/users` and `/channels` support simple text matching.

###### Images

Multipart forms will have to be used to post pictures. The following shows
a request to create a new image based release:

```http request
###
# POST image release form-data
POST http://localhost:8080/releases
Content-Type: multipart/form-data; boundary=WebAppBoundary

--WebAppBoundary
Content-Disposition: form-data; name=JSON
Content-Type: application/json

{
  "ownerChannel": "ownerChannel",
  "type": "image",
  "content": "content is only used for text type releases",
  "metadata": {
    "title": "title",
    "releaseDate": "2016-01-04T23:06:16.0175845+03:00",
    "genreDefining": "Example",
    "description": "multipart examole",
    "other": {
      "authors": [
        "aboutTime member"
      ],
      "genres": [
        "other", "genres"
      ]
    }
  }
}
--WebAppBoundary
Content-Disposition: form-data; name="image"; filename="image file name.jpg"
Content-Type: image/jpeg

< C:\path\to\image.jpg
--WebAppBoundary--
```

Note: the JSON part of the request is used to post metadata.


###### Auth

The following route is used to get JWT authorization tokens.

```http request
POST http://localhost:8080/token-auth
Content-Type: application/json

{
  "username": "rembrandt",
  "password": "password"
}
```

All username currently have the same password as the one shown.

JWT tokens will have to be attached using Bearer headers.
They are currently configured to expire in 15 minutes but they
 can be refreshed using the following route
 
 ```http request
GET http://localhost:8080/token-auth-refresh
Authorization: Bearer {{auth_token}}
 ```
### Thank You 

## Group members

```
{
Hanna Girma;
Beza Tsegaye; 
Yophthahe Amare;
Bilen Gizachew;
Michael Zewdu
}
```