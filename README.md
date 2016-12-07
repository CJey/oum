# OUM

An OpenVPN User Management utility

1. use oum to generate openvpn tls/password mode server configuration quicklly
1. use oum to generate a pattern script, then manage your iptables more easilly
1. use oum to guide user download openvpn client and configuration file easilly by a simple web(chinese support only)
1. use oum as openvpn server configuration hook, then verify user, assign ip by oum
1. oum support 8 digits otp as user authentication method, and this is default
1. oum support ip whitelist, domain whitelist, city whitelist to restrict user login

## Compile

Depends: linux, golang, git, build-essential, INTERNET(first)

Compile: ./build oum

Run: ./bin/oum

Install: cp bin/oum /usr/local/bin/

## Summary

Begin with oum, just follow `oum guide`

## Base

* oum help `[<command>...]`

    Show help.

* oum version

    Show Version Info

## User Management

Default, user data will stored at /var/lib/oum/oum.db, a sqlite3 db

Oum encourage using OTP Code to verify

* oum add `[<flags>] <username> [<password>]`

    Add user/device

* oum set `[<flags>] [<username>]`

    Update user config

* oum ifconfig `[<flags>] <username> [<config pair>...]`

    Assign/Show static ip and dns to user

* oum show `[<username>]`

    Show user config

* oum del `<username>`

    Delete user/device

* oum list `[<usernames>...]`

    List users/devices info

* oum enable `<username>`

    Enable user/device

* oum disable  `<username>`

    Disable user/device

* oum reset `<username>`

    Reset user password

* oum reconnect `<username>`

    Reconnect user/device

## OpenVPN hook

Used at openvpn server configuration, like this:

```bash
# use setenv to control oum
#setenv oum_sameip   1296000 # 15days
#setenv oum_samecity 604800 # 7days
#setenv oum_gateway  192.168.94.1
#setenv oum_dns      192.168.94.1 # support csv

up '/usr/local/bin/oum hook'
auth-user-pass-verify '/usr/local/bin/oum hook' via-env
client-connect '/usr/local/bin/oum hook'
client-disconnect '/usr/local/bin/oum hook'
down '/usr/local/bin/oum hook'
```

* oum hook `[<args>...]`

    Working as openvpn hook scripts

## Serving

By config `oum serv`, oum can disconnect connection of device, and distribute openvpn client configuration by `oum web`

* oum serv add `<conffile>`

    Add new server configuration file to serve

* oum serv del `<dev>`

    Delete a specified server configuration file to serve

* oum serv list

    List all server configuration file that oum serving now

## Helper

All subcommand below are optional, but sometimes useful

* oum guide

    Show general operation guide

    If you don't know how to begin, just follow this

* oum pattern `--out=OUT [<flags>]`

    Show pattern openvpn server configuration file

    If you hate to create openvpn configuration by youself, just use `oum pattern` to generate configuration file simply!

* oum iptables `--out=OUT`

    Output iptables pattern

    If you could not control iptables well, just use this pattern!

    Without correct iptables config, maybe the vpn could not work as you thought

* oum web `[<flags>]`

    Run http service to serve user to download config file

    Only support chinese now

* oum ipset `[<flags>] <sets>...`

    Resolve dns to update ipset

    This is very usefull to control traffic from/to domain(ddns normally)

* oum table `[<names>...]`

    Show sqlite3 table definition, you can hack db with utils sqlite3 by your hand

* oum verify `<username> <password>`

    Verify user, check password ok or no
