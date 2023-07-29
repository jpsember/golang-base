# Using self-signed keys to serve https on a local machine

For development purposes, it is convenient to be able to serve https on a local machine, without the complexity of getting a 'proper' https certificate or domain name registration.

## Generating a key pair

Run this command to generate a certificate and private key.  Replace all occurrences
of `URL` with an expression such as `foo.com`:

```
openssl req -x509 -days 3650 -out URL.crt -keyout URL.key \
  -newkey rsa:2048 -nodes -sha256 \
  -subj '/CN=URL.org' -extensions EXT -config <( \
   printf "[dn]\nCN=URL.org\n[req]\ndistinguished_name = dn\n[EXT]\nsubjectAltName=DNS:URL.org\nkeyUsage=digitalSignature\nextendedKeyUsage=serverAuth")
```


## Registering certificate with local machine

We'll assume a recent version of OSX.

1. Double click on the `crt` certificate file generated in the previous section.  This should open the Keychain Access program.

2. It will have added the certificate to a list.  Right-click on it, and select "Get Info".

3. Expand the 'Trust' tab near the top.

4. Change "When using this certificate" to "Always Trust".

5. Close the window, and type your password when it asks for it.



## Pointing the url to localhost

When you view your webserver in a browser (or curl) on your local machine, you want it to look at localhost.

In Linux, `/etc/hosts` is a file used by the operating system to translate hostnames to IP-addresses. It is also called the "hosts" file. By adding lines to this file, we can map arbitrary hostnames to arbitrary IP-addresses, which then we can use for testing websites locally.

Type:
```
sudo vi /etc/hosts
```
(Enter your password if prompted.)

Add this line to the end of the file:
```
127.0.0.1 foo.com
```

Replace `foo.com` with the `URL` value you used in the previous section.

Close the editor, saving changes.




## Using the certificate and private key in Go


Add the location of the certificate and private key to the appropriate line, e.g.:

```
err := http.ListenAndServeTLS(":443", "...path to foo.com.crt" , "...path to foo.com.key", nil)
```



# Notes

[This was helpful.](https://letsencrypt.org/docs/certificates-for-localhost/#making-and-trusting-your-own-certificates)

Also, BIG thanks to [Boris Reitman](https://www.linkedin.com/in/boris-reitman-8b2027134/)!


