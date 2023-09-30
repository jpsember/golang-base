Following instructions at https://certbot.eff.org/instructions?ws=other&os=ubuntufocal


```
~} sudo apt-get remove certbot
[sudo] password for jeff:
Reading package lists... Done
Building dependency tree
Reading state information... Done
Package 'certbot' is not installed, so not removed
0 upgraded, 0 newly installed, 0 to remove and 11 not upgraded.
~}
```

```
~} sudo snap install --classic certbot
certbot 2.6.0 from Certbot Project (certbot-eff✓) installed
~}
```

```
~} sudo ln -s /snap/bin/certbot /usr/bin/certbot
~}
```


...purchased a domain name...


```
~} nslookup pawsforaid.org
Server:   127.0.0.53
Address:  127.0.0.53#53

Non-authoritative answer:
Name: pawsforaid.org
Address: 172.232.171.126

~} sudo certbot certonly --standalone
Saving debug log to /var/log/letsencrypt/letsencrypt.log
Please enter the domain name(s) you would like on your certificate (comma and/or
space separated) (Enter 'c' to cancel): pawsforaid.org
Requesting a certificate for pawsforaid.org

Successfully received certificate.
Certificate is saved at: /etc/letsencrypt/live/pawsforaid.org/fullchain.pem
Key is saved at:         /etc/letsencrypt/live/pawsforaid.org/privkey.pem
This certificate expires on 2023-12-28.
These files will be updated when the certificate renews.
Certbot has set up a scheduled task to automatically renew this certificate in the background.

- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
If you like Certbot, please consider supporting our work by:
 * Donating to ISRG / Let's Encrypt:   https://letsencrypt.org/donate
 * Donating to EFF:                    https://eff.org/donate-le
- - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - -
~}
```

