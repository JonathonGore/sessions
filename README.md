# Sessions

* This implementation is based on the `sync` packages Map type.
* Thread safe map, that is most performant for a read heavy workload.

Create a session manager by providing a cookie name and an expiration length:
```
sm, err := NewManager("cookiename", 3600*24*365)
if err != nil {
    return nil, err
}
```
