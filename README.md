# OUM

An OpenVPN User Management utility

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
up '/usr/local/bin/oum hook'
auth-user-pass-verify '/usr/local/bin/oum hook --sameip 1296000 --samecity 604800' via-env
client-connect '/usr/local/bin/oum hook --gateway 192.168.94.1 --dns 192.168.94.1'
client-disconnect '/usr/local/bin/oum hook'
down '/usr/local/bin/oum hook'
```

* oum hook `[<flags>] [<args>...]`

    Working as openvpn hook scripts

## Serving

By config `oum serv`, oum can disconnect connection of device, and distrbute openvpn client configuration by `oum web`

* oum serv add `<conffile>`

    Add new server configuration file to serve

* oum serv del `<dev>`

    Delete a specified server configuration file to serve

* oum serv list

    List all server configuration file that oum serving now

## Helper

All subcommand below are optional, but sometimes usefull

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
