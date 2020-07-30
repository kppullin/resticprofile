[![Build Status](https://travis-ci.com/creativeprojects/resticprofile.svg?branch=master)](https://travis-ci.com/creativeprojects/resticprofile)
[![Go Report Card](https://goreportcard.com/badge/github.com/creativeprojects/resticprofile)](https://goreportcard.com/report/github.com/creativeprojects/resticprofile)

# resticprofile
Configuration profiles manager for [restic backup](https://restic.net/)

**resticprofile** is the missing link between a configuration file and restic backup. Creating a configuration file for restic has been [discussed before](https://github.com/restic/restic/issues/16), but seems to be a very low priority right now.

With resticprofile:

* You no longer need to remember command parameters and environment variables
* You can create multiple profiles inside one configuration file
* A profile can inherit all the options from another profile
* You can run the forget command before or after a backup (in a section called *retention*)
* You can check a repository before or after a backup
* You can create groups of profiles that will run sequentially
* You can run shell commands before or after running a profile: useful if you need to mount and unmount your backup disk for example
* You can run a shell command if an error occurred (at any time)
* You can send a backup stream via _stdin_
* You can start restic at a lower or higher priority (Priority Class in Windows, *nice* in all unixes) and/or _ionice_ (only available on Linux)
* It can check that you have enough memory before starting a backup. (I've had some backups that literally killed a server with swap disabled)
* **[new for v0.9.0]** You can generate cryptographically secure random keys to use as a restic key file
* **[new for v0.9.0]** You can easily schedule backups, retentions and checks (works for *systemd*, *launchd* and *windows task scheduler*)

The configuration file accepts various formats:
* [TOML](https://github.com/toml-lang/toml) : configuration file with extension _.toml_ and _.conf_ to keep compatibility with versions before 0.6.0
* [JSON](https://en.wikipedia.org/wiki/JSON) : configuration file with extension _.json_
* [YAML](https://en.wikipedia.org/wiki/YAML) : configuration file with extension _.yaml_
* [HCL](https://github.com/hashicorp/hcl): configuration file with extension _.hcl_

For the rest of the documentation, I'll be mostly showing examples using the TOML file configuration format (because it was the only one supported before version 0.6.0) but you can pick your favourite: they all work with resticprofile.

# Table of Contents

* [resticprofile](#resticprofile)
* [Table of Contents](#table-of-contents)
  * [Requirements](#requirements)
  * [Installation (macOS, Linux &amp; other unixes)](#installation-macos-linux--other-unixes)
    * [Installation for Windows using bash](#installation-for-windows-using-bash)
    * [Manual installation (Windows)](#manual-installation-windows)
  * [Upgrade](#upgrade)
  * [Using docker image](#using-docker-image)
    * [Container host name](#container-host-name)
  * [Configuration format](#configuration-format)
  * [Configuration examples](#configuration-examples)
  * [Configuration paths](#configuration-paths)
    * [macOS X](#macos-x)
    * [Other unixes (Linux and BSD)](#other-unixes-linux-and-bsd)
    * [Windows](#windows)
  * [Path resolution in configuration](#path-resolution-in-configuration)
  * [Run commands before, after success or after failure](#run-commands-before-after-success-or-after-failure)
  * [Using resticprofile](#using-resticprofile)
  * [Command line reference](#command-line-reference)
  * [Minimum memory required](#minimum-memory-required)
  * [Generating random keys](#generating-random-keys)
  * [Scheduled backups](#scheduled-backups)
    * [Schedule configuration](#schedule-configuration)
      * [schedule\-permission](#schedule-permission)
      * [schedule\-log](#schedule-log)
      * [schedule](#schedule)
    * [Scheduling commands](#scheduling-commands)
      * [Examples of scheduling commands under Windows](#examples-of-scheduling-commands-under-windows)
      * [Examples of scheduling commands under Linux](#examples-of-scheduling-commands-under-linux)
      * [Examples of scheduling commands under macOS](#examples-of-scheduling-commands-under-macos)
    * [Changing schedule\-permission from user to system, or system to user](#changing-schedule-permission-from-user-to-system-or-system-to-user)
  * [Configuration file reference](#configuration-file-reference)
  * [Appendix](#appendix)
  * [Using resticprofile and systemd](#using-resticprofile-and-systemd)
    * [systemd calendars](#systemd-calendars)
    * [First time schedule](#first-time-schedule)
  * [Using resticprofile and launchd on macOS](#using-resticprofile-and-launchd-on-macos)
    * [User agent](#user-agent)
      * [Special case of schedule\-permission=user with sudo](#special-case-of-schedule-permissionuser-with-sudo)
    * [Daemon](#daemon)


## Requirements

Since version 0.6.0, resticprofile no longer needs python installed on your machine. It is distributed as an executable (same as restic).

It's been actively tested on macOS X and Linux, and regularly tested on Windows.

**This is at _beta_ stage. Please avoid using it in production. Or at least test carefully first. Even though I'm using it on my servers, I cannot guarantee all combinations of configuration are going to work properly for you.**

## Installation (macOS, Linux & other unixes)

Here's a simple script to download the binary automatically. It works on mac OS X, FreeBSD, OpenBSD and Linux:

```
$ curl -sfL https://raw.githubusercontent.com/creativeprojects/resticprofile/master/install.sh | sh
```

It should copy resticprofile in a `bin` directory under your current directory.

If you need more control, you can save the shell script and run it manually:

```
$ curl -LO https://raw.githubusercontent.com/creativeprojects/resticprofile/master/install.sh
$ chmod +x install.sh
$ sudo ./install.sh -b /usr/local/bin
```

It will install resticprofile in `/usr/local/bin/`


### Installation for Windows using bash

You can use the same script if you're using bash in Windows (via WSL, git bash, etc.)

```
$ curl -LO https://raw.githubusercontent.com/creativeprojects/resticprofile/master/install.sh
$ ./install.sh
```
It will create a `bin` directory under your current directory and place `resticprofile.exe` in it.

### Manual installation (Windows)

- Download the package corresponding to your system and CPU from the [release page](https://github.com/creativeprojects/resticprofile/releases)
- Once downloaded you need to open the archive and copy the binary file `resticprofile` (or `resticprofile.exe`) in your PATH.

## Upgrade

Once installed, you can easily upgrade resticprofile to the latest release using this command:

```
$ resticprofile self-update
```

Versions 0.6.x were using a flag instead:

```
$ resticprofile --self-update
```

_Please note there's an issue with self-updating from linux with ARM processors (like a raspberry pi)_

## Using docker image ##

You can run resticprofile inside a docker container. It is probably the easiest way to install resticprofile (and restic at the same time) and keep it updated.

**But** be aware that you will need to mount your backup source (and destination if it's local) as a docker volume.
Depending on your operating system, the backup might be **slower**. Volumes mounted on a mac OS host are well known for being quite slow.

By default, the resticprofile container starts at `/resticprofile`. So you can feed a configuration this way:

```
$ docker run -it --rm -v $PWD/examples:/resticprofile creativeprojects/resticprofile
```

You can list your profiles:
```
$ docker run -it --rm -v $PWD/examples:/resticprofile creativeprojects/resticprofile profiles
```

### Container host name

Each time a container is started, it gets assigned a new random name. You should probably force a hostname to your container...

```
$ docker run -it --rm -v $PWD:/resticprofile -h my-machine creativeprojects/resticprofile -n profile backup
```

... or in your configuration:

```ini
[profile]
host = "my-machine"
```


## Configuration format

* A configuration is a set of _profiles_.
* Each profile is in its own `[section]`.
* Inside each profile, you can specify different flags for each command.
* A command definition is `[section.command]`.

All the restic flags can be defined in a section. For most of them you just need to remove the two dashes in front.

To set the flag `--password-file password.txt` you need to add a line like
```
password-file = "password.txt"
```

There's **one exception**: the flag `--repo` is named `repository` in the configuration

Let's say you normally use this command:

```
restic --repo "local:/backup" --password-file "password.txt" --verbose backup /home
```

For resticprofile to generate this command automatically for you, here's the configuration file:

```ini
[default]
repository = "local:/backup"
password-file = "password.txt"

[default.backup]
verbose = true
source = [ "/home" ]
```

You may have noticed the `source` flag is accepting an array of values (inside brackets)

Now, assuming this configuration file is named `profiles.conf` in the current folder, you can simply run

```
resticprofile backup
```

## Configuration examples

Here's a simple configuration file using a Microsoft Azure backend:

```ini
[default]
repository = "azure:restic:/"
password-file = "key"

[default.env]
AZURE_ACCOUNT_NAME = "my_storage_account"
AZURE_ACCOUNT_KEY = "my_super_secret_key"

[default.backup]
exclude-file = "excludes"
exclude-caches = true
one-file-system = true
tag = [ "root" ]
source = [ "/", "/var" ]
```

Here's a more complex configuration file showing profile inheritance and two backup profiles using the same repository:

```ini
[global]
# ionice is available on Linux only
ionice = false
ionice-class = 2
ionice-level = 6
# priority is using priority class on windows, and "nice" on unixes - it's acting on CPU usage only
priority = "low"
# run 'snapshots' when no command is specified when invoking resticprofile
default-command = "snapshots"
# initialize a repository if none exist at location
initialize = false
# resticprofile won't start a profile if there's less than 100MB of RAM available
min-memory = 100

# a group is a profile that will call all profiles one by one
[groups]
# when starting a backup on profile "full-backup", it will run the "root" and "src" backup profiles
full-backup = [ "root", "src" ]

# Default profile when not specified (-n or --name)
# Please note there's no default inheritance from the 'default' profile (you can use the 'inherit' flag if needed)
[default]
# you can use a relative path, it will be relative to the configuration file
repository = "/backup"
password-file = "key"
initialize = false
# will run these scripts before and after each command (including 'backup')
run-before = "mount /backup"
run-after = "umount /backup"
# if a restic command fails, the run-after won't be running
# add this parameter to run the script in case of a failure
run-after-fail = "umount /backup"

[default.env]
TMPDIR= "/tmp"

[no-cache]
inherit = "default"
no-cache = true
initialize = false

# New profile named 'root'
[root]
inherit = "default"
initialize = true
# this will add a LOCAL lockfile so you cannot run the same profile more than once at a time
# (it's totally independent of the restic locks on the repository)
lock = "/tmp/resticprofile-root.lock"

# 'backup' command of profile 'root'
[root.backup]
# files with no path are relative to the configuration file
exclude-file = [ "root-excludes", "excludes" ]
exclude-caches = true
one-file-system = false
tag = [ "test", "dev" ]
source = [ "/" ]
# if scheduled, will run every dat at midnight
schedule = "daily"
schedule-permission = "system"
# run this after a backup to share a repository between a user and root (via sudo)
run-after = "chown -R $SUDO_USER $HOME/.cache/restic /backup"

# retention policy for profile root
[root.retention]
before-backup = false
after-backup = true
keep-last = 3
keep-hourly = 1
keep-daily = 1
keep-weekly = 1
keep-monthly = 1
keep-yearly = 1
keep-within = "3h"
keep-tag = [ "forever" ]
compact = false
prune = false
# if path is NOT specified, it will be copied from the 'backup' source
# path = []
# the tags are NOT copied from the 'backup' command
tag = [ "test", "dev" ]
# host can be a boolean ('true' meaning current hostname) or a string to specify a different hostname
host = true

# New profile named 'src'
[src]
inherit = "default"
initialize = true

# 'backup' command of profile 'src'
[src.backup]
exclude = [ '/**/.git' ]
exclude-caches = true
one-file-system = false
tag = [ "test", "dev" ]
source = [ "./src" ]
check-before = true
# will only run these scripts before and after a backup
run-before = [ "echo Starting!", "ls -al ./src" ]
run-after = "sync"
# if scheduled, will run every 30 minutes
schedule = "*:0,30"
schedule-permission = "user"

# retention policy for profile src
[src.retention]
before-backup = false
after-backup = true
keep-within = "30d"
compact = false
prune = true

# check command of profile src
[src.check]
read-data = true
# if scheduled, will check the repository the first day of each month at 3am
schedule = "*-*-01 03:00"

```

And another simple example for Windows:

```ini
[global]
restic-binary = "c:\\ProgramData\\chocolatey\\bin\\restic.exe"

# Default profile when not specified (-n or --name)
# Please note there's no default inheritance from the 'default' profile (you can use the 'inherit' flag if needed)
[default]
repository = "local:r:/"
password-file = "key"
initialize = false

# New profile named 'test'
[test]
inherit = "default"
initialize = true

# 'backup' command of profile 'test'
[test.backup]
tag = [ "windows" ]
source = [ "c:\\" ]
check-after = true
run-before = "dir /l"
run-after = "echo All Done!"

```

Simple example sending a file via stdin

```ini

[stdin]
repository = "local:/backup/restic"
password-file = "key"

[stdin.backup]
stdin = true
stdin-filename = "stdin-test"
tag = [ 'stdin' ]

```

## Configuration paths

The default name for the configuration file is `profiles`, without an extension.
You can change the name and its path with the `--config` or `-c` option on the command line.
You can set a specific extension `-c profiles.conf` to load a TOML format file.
If you set a filename with no extension instead, resticprofile will load the first file it finds with any of these extensions:
- .conf (toml format)
- .yaml
- .toml
- .json
- .hcl

### macOS X

resticprofile will search for your configuration file in these folders:
- _current directory_
- ~/Library/Preferences/resticprofile/
- /Library/Preferences/resticprofile/
- /usr/local/etc/
- /usr/local/etc/restic/
- /usr/local/etc/resticprofile/
- /etc/
- /etc/restic/
- /etc/resticprofile/
- /opt/local/etc/
- /opt/local/etc/restic/
- /opt/local/etc/resticprofile/
- ~/ ($HOME directory)

### Other unixes (Linux and BSD)

resticprofile will search for your configuration file in these folders:
- _current directory_
- ~/.config/resticprofile/
- /etc/xdg/resticprofile/
- /usr/local/etc/
- /usr/local/etc/restic/
- /usr/local/etc/resticprofile/
- /etc/
- /etc/restic/
- /etc/resticprofile/
- /opt/local/etc/
- /opt/local/etc/restic/
- /opt/local/etc/resticprofile/
- ~/ ($HOME directory)

### Windows

resticprofile will search for your configuration file in these folders:
- _current directory_
- %USERPROFILE%\AppData\Local\
- c:\ProgramData\
- c:\restic\
- c:\resticprofile\
- %USERPROFILE%\


## Path resolution in configuration

All files path in the configuration are resolved from the configuration path. The big **exception** being `source` in `backup` section where it's resolved from the current path where you started resticprofile.

## Run commands before, after success or after failure

resticprofile has 2 places you can run commands around restic:

- commands that will run before and after every restic command (snapshots, backup, check, forget, prune, mount, etc.). These are placed at the root of each profile.
- commands that will only run before and after a backup: these are placed in the backup section of your profiles.

Here's an example of all the commands that you can run in a profile:

```yaml
documents:
    inherit: default
    run-before: "echo == run-before profile $PROFILE_NAME command $PROFILE_COMMAND"
    run-after: "echo == run-after profile $PROFILE_NAME command $PROFILE_COMMAND"
    run-after-fail: "echo == Error in profile $PROFILE_NAME command $PROFILE_COMMAND: $ERROR"
    backup:
        run-before: "echo === run-before backup profile $PROFILE_NAME command $PROFILE_COMMAND"
        run-after: "echo === run-after backup profile $PROFILE_NAME command $PROFILE_COMMAND"
        source: ~/Documents
```

`run-before`, `run-after` and `run-after-fail` can be a string, or an array of strings if you need to run more than one command

A few environment variables will be set before running these commands:
- `PROFILE_NAME`
- `PROFILE_COMMAND`: backup, check, forget, etc.

Additionally for the `run-after-fail` commands, the `ERROR` environment variable will be set to the latest error message.

## Using resticprofile

Here are a few examples how to run resticprofile (using the main example configuration file)

See all snapshots of your `[default]` profile:

```
$ resticprofile
```

See all available profiles in your configuration file (and the restic commands where some flags are defined):

```
$ resticprofile profiles

Profiles available:
  stdin:     (backup)
  default:   (env)
  root:      (retention, backup)
  src:       (retention, backup)
  linux:     (retention, backup, snapshots, env)
  no-cache:  (n/a)

Groups available:
  full-backup:  root, src

```

Backup root & src profiles (using _full-backup_ group shown earlier)

```
$ resticprofile --name "full-backup" backup
```

Assuming the _stdin_ profile from the configuration file shown before, the command to send a mysqldump to the backup is as simple as:

```
$ mysqldump --all-databases | resticprofile --name stdin backup
```

Mount the default profile (_default_) in /mnt/restic:

```
$ resticprofile mount /mnt/restic
```

Display quick help

```
$ resticprofile --help

Usage of resticprofile:
	resticprofile [resticprofile flags] [command] [restic flags]

resticprofile flags:
  -c, --config string   configuration file (default "profiles")
      --dry-run         display the restic commands instead of running them
  -f, --format string   file format of the configuration (default is to use the file extension)
  -h, --help            display this help
  -l, --log string      logs into a file instead of the console
  -n, --name string     profile name (default "default")
      --no-ansi         disable ansi control characters (disable console colouring)
  -q, --quiet           display only warnings and errors
      --theme string    console colouring theme (dark, light, none) (default "light")
  -v, --verbose         display all debugging information
  -w, --wait            wait at the end until the user presses the enter key

resticprofile own commands:
   self-update   update resticprofile to latest version (does not update restic)
   profiles      display profile names from the configuration file
   show          show all the details of the current profile
   random-key    generate a cryptographically secure random key to use as a restic key file
   schedule      schedule a backup
   unschedule    remove a scheduled backup
   status        display the status of a scheduled backup job


```

A command is either a restic command or a resticprofile own command.


## Command line reference ##

There are not many options on the command line, most of the options are in the configuration file.

* **[-h]**: Display quick help
* **[-c | --config] configuration_file**: Specify a configuration file other than the default
* **[-f | --format] configuration_format**: Specify the configuration file format: `toml`, `yaml`, `json` or `hcl`
* **[-n | --name] profile_name**: Profile section to use from the configuration file
* **[--dry-run]**: Doesn't run the restic command but display the command line instead
* **[-q | --quiet]**: Force resticprofile and restic to be quiet (override any configuration from the profile)
* **[-v | --verbose]**: Force resticprofile and restic to be verbose (override any configuration from the profile)
* **[--no-ansi]**: Disable console colouring (to save output into a log file)
* **[--theme]**: Can be `light`, `dark` or `none`. The colours will adjust to a 
light or dark terminal (none to disable colouring)
* **[-l | --log] log_file**: To write the logs in file instead of displaying on the console
* **[-w | --wait]**: Wait at the very end of the execution for the user to press enter. This is only useful in Windows when resticprofile is started from explorer and the console window closes automatically at the end.
* **[resticprofile OR restic command]**: Like snapshots, backup, check, prune, forget, mount, etc.
* **[additional flags]**: Any additional flags to pass to the restic command line

## Minimum memory required

restic can be memory hungry. I'm running a few servers with no swap (I know: it is _bad_) and I managed to kill some of them during a backup.
For that matter I've introduced a parameter in the `global` section called `min-memory`. The **default value is 100MB**. You can disable it by using a value of `0`.

It compares against `(total - used)` which is probably the best way to know how much memory is available (that is including the memory used for disk buffers/cache).

## Generating random keys

resticprofile has a handy tool to generate cryptographically secure random keys encoded in base64. You can simply put this key into a file and use it as a strong key for restic

On Linux and FreeBSD, the generator uses getrandom(2) if available, /dev/urandom otherwise. On OpenBSD, the generator uses getentropy(2). On other Unix-like systems, the generator reads from /dev/urandom. On Windows systems, the generator uses the CryptGenRandom API. On Wasm, the generator uses the Web Crypto API. 
[Reference from the Go documentation](https://golang.org/pkg/crypto/rand/#pkg-variables)

```
$ resticprofile random-key
```

generates a 1024 bytes random key (converted into 1368 base64 characters) and displays it on the console

To generate a different size of key, you can specify the bytes length on the command line:

```
$ resticprofile random-key 2048
```

## Scheduled backups

resticprofile is capable of managing scheduled backups for you:
- using **systemd** where available (Linux and various unixes)
- using **launchd** on macOS X
- using **Task Scheduler** on Windows

Each profile can be scheduled independently (groups are not available for scheduling yet).

These 3 profile sections are accepting a schedule configuration:
- backup
- retention (when not run before or after a backup)
- check

which mean you can schedule backup, retention (`forget` command) and repository check independently (I recommend to use a local `lock` in this case).

### Schedule configuration

The schedule configuration consists of a few parameters which can be added on each profile:

```ini
[profile.backup]
schedule = "*:00,30"
schedule-permission = "system"
schedule-log = "profile-backup.log"
```



#### schedule-permission

`schedule-permission` accepts two parameters: `user` or `system`:

* `user`: your backup will be running using your current user permissions on files. That's probably what you want if you're only saving your documents (or any other file inside your profile).

* `system`: if you need to access some system or protected files. You will need to run resticprofile with `sudo` on unixes and with elevated prompt on Windows (please note on Windows resticprofile will ask you for elevated permissions automatically if needed)

* *empty*: resticprofile will try its best guess based on how you started it (with sudo or as a normal user) and fallback to `user`

#### schedule-log

Allow to redirect all output from resticprofile and restic to a file

#### schedule

The `schedule` parameter accepts many forms of input from the [systemd calendar event](https://www.freedesktop.org/software/systemd/man/systemd.time.html#Calendar%20Events) type. This is by far the easiest to use: **It is the same format used to schedule on macOS and Windows**.

The most general form is:
```
weekdays year-month-day hour:minute:second
```

- use `*` to mean any
- use `,` to separate multiple entries
- use `..` for a range

**limitations**:
- the divider (`/`), the `~` and timezones are not (yet?) supported on macOS and Windows.
- the `year` and `second` fields have no effect on macOS. They do have limited availability on Windows (they don't make much sense anyway).

Here are a few examples (taken from the systemd documentation):

```
On the left is the user input, on the right is the full format understood by the system

  Sat,Thu,Mon..Wed,Sat..Sun → Mon..Thu,Sat,Sun *-*-* 00:00:00
      Mon,Sun 12-*-* 2,1:23 → Mon,Sun 2012-*-* 01,02:23:00
                    Wed *-1 → Wed *-*-01 00:00:00
           Wed..Wed,Wed *-1 → Wed *-*-01 00:00:00
                 Wed, 17:48 → Wed *-*-* 17:48:00
Wed..Sat,Tue 12-10-15 1:2:3 → Tue..Sat 2012-10-15 01:02:03
                *-*-7 0:0:0 → *-*-07 00:00:00
                      10-15 → *-10-15 00:00:00
        monday *-12-* 17:00 → Mon *-12-* 17:00:00
     Mon,Fri *-*-3,1,2 *:30 → Mon,Fri *-*-01,02,03 *:30:00
       12,14,13,12:20,10,30 → *-*-* 12,13,14:10,20,30:00
            12..14:10,20,30 → *-*-* 12..14:10,20,30:00
                03-05 08:05 → *-03-05 08:05:00
                      05:40 → *-*-* 05:40:00
        Sat,Sun 12-05 08:05 → Sat,Sun *-12-05 08:05:00
              Sat,Sun 08:05 → Sat,Sun *-*-* 08:05:00
           2003-03-05 05:40 → 2003-03-05 05:40:00
             2003-02..04-05 → 2003-02..04-05 00:00:00
                 2003-03-05 → 2003-03-05 00:00:00
                      03-05 → *-03-05 00:00:00
                     hourly → *-*-* *:00:00
                      daily → *-*-* 00:00:00
                    monthly → *-*-01 00:00:00
                     weekly → Mon *-*-* 00:00:00
                     yearly → *-01-01 00:00:00
                   annually → *-01-01 00:00:00
```

The `schedule` can be a string or an array of string (to allow for multiple schedules)

Here's an example of a YAML configuration:

```yaml
default:
    repository: "d:\\backup"
    password-file: key

self:
    inherit: default
    backup:
        source: "."
        schedule:
        - "Mon..Fri *:00,15,30,45" # every 15 minutes on weekdays
        - "Sat,Sun 0,12:00"        # twice a day on week-ends
        schedule-permission: user
    retention:
        schedule: "sun 3:30"
        schedule-permission: user
```

### Scheduling commands

resticprofile accepts these internal commands:
- schedule
- unschedule
- status

Please note the display of the `status` command will be OS dependant.

#### Examples of scheduling commands under Windows

If you create a task with `user` permission under Windows, you will need to enter your password to validate the task. It's a requirement of the task scheduler. I'm inviting you to review the code to make sure I'm not emailing your password to myself. Seriously you shouldn't trust anyone.

Example of the `schedule` command under Windows (with git bash):

```
$ resticprofile -c examples/windows.yaml -n self schedule

Analyzing backup schedule 1/2
=================================
  Original form: Mon..Fri *:00,15,30,45
Normalized form: Mon..Fri *-*-* *:00,15,30,45:00
    Next elapse: Wed Jul 22 21:30:00 BST 2020
       (in UTC): Wed Jul 22 20:30:00 UTC 2020
       From now: 1m52s left

Analyzing backup schedule 2/2
=================================
  Original form: Sat,Sun 0,12:00
Normalized form: Sat,Sun *-*-* 00,12:00:00
    Next elapse: Sat Jul 25 00:00:00 BST 2020
       (in UTC): Fri Jul 24 23:00:00 UTC 2020
       From now: 50h31m52s left

Creating task for user Creative Projects
Task Scheduler requires your Windows password to validate the task: 

2020/07/22 21:28:15 scheduled job self/backup created

Analyzing retention schedule 1/1
=================================
  Original form: sun 3:30
Normalized form: Sun *-*-* 03:30:00
    Next elapse: Sun Jul 26 03:30:00 BST 2020
       (in UTC): Sun Jul 26 02:30:00 UTC 2020
       From now: 78h1m44s left

2020/07/22 21:28:22 scheduled job self/retention created
```

To see the status of the triggers, you can use the `status` command:

```
$ resticprofile -c examples/windows.yaml -n self status

Analyzing backup schedule 1/2
=================================
  Original form: Mon..Fri *:00,15,30,45
Normalized form: Mon..Fri *-*-* *:00,15,30,45:00
    Next elapse: Wed Jul 22 21:30:00 BST 2020
       (in UTC): Wed Jul 22 20:30:00 UTC 2020
       From now: 14s left

Analyzing backup schedule 2/2
=================================
  Original form: Sat,Sun 0,12:*
Normalized form: Sat,Sun *-*-* 00,12:*:00
    Next elapse: Sat Jul 25 00:00:00 BST 2020
       (in UTC): Fri Jul 24 23:00:00 UTC 2020
       From now: 50h29m46s left

           Task: \resticprofile backup\self backup
           User: Creative Projects
    Working Dir: D:\Source\resticprofile
           Exec: D:\Source\resticprofile\resticprofile.exe --no-ansi --config examples/windows.yaml --name self backup
        Enabled: true
          State: ready
    Missed runs: 0
  Last Run Time: 2020-07-22 21:30:00 +0000 UTC
    Last Result: 0
  Next Run Time: 2020-07-22 21:45:00 +0000 UTC

Analyzing retention schedule 1/1
=================================
  Original form: sun 3:30
Normalized form: Sun *-*-* 03:30:00
    Next elapse: Sun Jul 26 03:30:00 BST 2020
       (in UTC): Sun Jul 26 02:30:00 UTC 2020
       From now: 77h59m46s left

           Task: \resticprofile backup\self retention
           User: Creative Projects
    Working Dir: D:\Source\resticprofile
           Exec: D:\Source\resticprofile\resticprofile.exe --no-ansi --config examples/windows.yaml --name self forget
        Enabled: true
          State: ready
    Missed runs: 0
  Last Run Time: 1999-11-30 00:00:00 +0000 UTC
    Last Result: 267011
  Next Run Time: 2020-07-26 03:30:00 +0000 UTC

```

To remove the schedule, use the `unschedule` command:

```
$ resticprofile -c examples/windows.yaml -n self unschedule
2020/07/22 21:34:51 scheduled job self/backup removed
2020/07/22 21:34:51 scheduled job self/retention removed
```

#### Examples of scheduling commands under Linux

With this example of configuration for Linux:

```yaml

default:
    password-file: key
    repository: /tmp/backup

test1:
    inherit: default
    backup:
        source: ./
        schedule: "*:00,15,30,45"
        schedule-permission: user
    check:
        schedule: "*-*-1"
        schedule-permission: user

```

```
$ resticprofile -c examples/linux.yaml -n test1 schedule

Analyzing backup schedule 1/1
=================================
  Original form: *:00,15,30,45
Normalized form: *-*-* *:00,15,30,45:00
    Next elapse: Thu 2020-07-23 17:15:00 BST
       (in UTC): Thu 2020-07-23 16:15:00 UTC
       From now: 6min left

2020/07/23 17:08:51 writing /home/user/.config/systemd/user/resticprofile-backup@profile-test1.service
2020/07/23 17:08:51 writing /home/user/.config/systemd/user/resticprofile-backup@profile-test1.timer
Created symlink /home/user/.config/systemd/user/timers.target.wants/resticprofile-backup@profile-test1.timer → /home/user/.config/systemd/user/resticprofile-backup@profile-test1.timer.
2020/07/23 17:08:51 scheduled job test1/backup created

Analyzing check schedule 1/1
=================================
  Original form: *-*-1
Normalized form: *-*-01 00:00:00
    Next elapse: Sat 2020-08-01 00:00:00 BST
       (in UTC): Fri 2020-07-31 23:00:00 UTC
       From now: 1 weeks 1 days left

2020/07/23 17:08:51 writing /home/user/.config/systemd/user/resticprofile-check@profile-test1.service
2020/07/23 17:08:51 writing /home/user/.config/systemd/user/resticprofile-check@profile-test1.timer
Created symlink /home/user/.config/systemd/user/timers.target.wants/resticprofile-check@profile-test1.timer → /home/user/.config/systemd/user/resticprofile-check@profile-test1.timer.
2020/07/23 17:08:51 scheduled job test1/check created
```

The `status` command shows a combination of `journalctl` displaying errors (only) in the last month and `systemctl status`:

```
$ resticprofile -c examples/linux.yaml -n test1 status

Analyzing backup schedule 1/1
=================================
  Original form: *:00,15,30,45
Normalized form: *-*-* *:00,15,30,45:00
    Next elapse: Tue 2020-07-28 15:15:00 BST
       (in UTC): Tue 2020-07-28 14:15:00 UTC
       From now: 4min 44s left

-- Logs begin at Wed 2020-06-17 11:09:19 BST, end at Tue 2020-07-28 15:10:10 BST. --
Jul 27 20:48:01 Desktop76 systemd[2986]: Failed to start resticprofile backup for profile test1 in examples/linux.yaml.
Jul 27 21:00:55 Desktop76 systemd[2986]: Failed to start resticprofile backup for profile test1 in examples/linux.yaml.
Jul 27 21:15:34 Desktop76 systemd[2986]: Failed to start resticprofile backup for profile test1 in examples/linux.yaml.

● resticprofile-backup@profile-test1.timer - backup timer for profile test1 in examples/linux.yaml
   Loaded: loaded (/home/user/.config/systemd/user/resticprofile-backup@profile-test1.timer; enabled; vendor preset: enabled)
   Active: active (waiting) since Tue 2020-07-28 15:10:06 BST; 8s ago
  Trigger: Tue 2020-07-28 15:15:00 BST; 4min 44s left

Jul 28 15:10:06 Desktop76 systemd[2951]: Started backup timer for profile test1 in examples/linux.yaml.


Analyzing check schedule 1/1
=================================
  Original form: *-*-1
Normalized form: *-*-01 00:00:00
    Next elapse: Sat 2020-08-01 00:00:00 BST
       (in UTC): Fri 2020-07-31 23:00:00 UTC
       From now: 3 days left

-- Logs begin at Wed 2020-06-17 11:09:19 BST, end at Tue 2020-07-28 15:10:10 BST. --
Jul 27 19:39:59 Desktop76 systemd[2986]: Failed to start resticprofile check for profile test1 in examples/linux.yaml.

● resticprofile-check@profile-test1.timer - check timer for profile test1 in examples/linux.yaml
   Loaded: loaded (/home/user/.config/systemd/user/resticprofile-check@profile-test1.timer; enabled; vendor preset: enabled)
   Active: active (waiting) since Tue 2020-07-28 15:10:07 BST; 7s ago
  Trigger: Sat 2020-08-01 00:00:00 BST; 3 days left

Jul 28 15:10:07 Desktop76 systemd[2951]: Started check timer for profile test1 in examples/linux.yaml.


```

And `unschedule`:

```
$ resticprofile -c examples/linux.yaml -n test1 unschedule
Removed /home/user/.config/systemd/user/timers.target.wants/resticprofile-backup@profile-test1.timer.
2020/07/23 17:13:42 scheduled job test1/backup removed
Removed /home/user/.config/systemd/user/timers.target.wants/resticprofile-check@profile-test1.timer.
2020/07/23 17:13:42 scheduled job test1/check removed
```

#### Examples of scheduling commands under macOS

macOS has a very tight protection system when running scheduled tasks (also called agents).

Under macOS, resticprofile is asking if you want to start a profile right now so you can give the access needed to the task (it will consist on a few popup windows)

Here's an example of scheduling a backup to Azure (which needs network access):

```
% resticprofile -v -c examples/private/azure.yaml -n self schedule

Analyzing backup schedule 1/1
=================================
  Original form: *:0,15,30,45:00
Normalized form: *-*-* *:00,15,30,45:00
    Next elapse: Tue Jul 28 23:00:00 BST 2020
       (in UTC): Tue Jul 28 22:00:00 UTC 2020
       From now: 2m34s left


By default, a macOS agent access is restricted. If you leave it to start in the background it's likely to fail.
You have to start it manually the first time to accept the requests for access:

% launchctl start local.resticprofile.self.backup

Do you want to start it now? (Y/n):
2020/07/28 22:57:26 scheduled job self/backup created
```

Right after you started the profile, you should get some popup asking you to grant access to various files/folders/network.

If you backup your files to an external repository on a network, you should get this popup window:

!["resticprofile" would like to access files on a network volume](https://github.com/creativeprojects/resticprofile/raw/master/network_volume.png)


### Changing schedule-permission from user to system, or system to user

If you need to change the permission of a schedule, **please be sure to `unschedule` the profile before**.

This order is important:

- `unschedule` the job first. resticprofile does **not** keep track of how your profile **was** installed, so you have to remove the schedule first
- now you can change your permission (`user` to `system`, or `system` to `user`)
- `schedule` your updated profile

## Configuration file reference

`[global]`

`global` is a fixed name

None of these flags are passed on the restic command line

* **ionice**: true / false
* **ionice-class**: integer
* **ionice-level**: integer
* **nice**: true / false OR integer
* **priority**: string = `Idle`, `Background`, `Low`, `Normal`, `High`, `Highest`
* **default-command**: string
* **initialize**: true / false
* **restic-binary**: string
* **min-memory**: integer (MB)

`[profile]`

`profile` is the name of your profile

Flags used by resticprofile only

* ****inherit****: string
* **initialize**: true / false
* **lock**: string: specify a local lockfile
* **run-before**: string OR list of strings
* **run-after**: string OR list of strings
* **run-after-fail**: string OR list of strings

Flags passed to the restic command line

* **cacert**: string
* **cache-dir**: string
* **cleanup-cache**: true / false
* **json**: true / false
* **key-hint**: string
* **limit-download**: integer
* **limit-upload**: integer
* **no-cache**: true / false
* **no-lock**: true / false
* **option**: string OR list of strings
* **password-command**: string
* **password-file**: string
* **quiet**: true / false
* **repository**: string **(will be passed as 'repo' to the command line)**
* **tls-client-cert**: string
* **verbose**: true / false OR integer

`[profile.backup]`

Flags used by resticprofile only

* **run-before**: string OR list of strings
* **run-after**: string OR list of strings
* **check-before**: true / false
* **check-after**: true / false
* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-log**: string

Flags passed to the restic command line

* **exclude**: string OR list of strings
* **exclude-caches**: true / false
* **exclude-file**: string OR list of strings
* **exclude-if-present**: string OR list of strings
* **files-from**: string OR list of strings
* **force**: true / false
* **host**: true / false OR string
* **iexclude**: string OR list of strings
* **ignore-inode**: true / false
* **one-file-system**: true / false
* **parent**: string
* **stdin**: true / false
* **stdin-filename**: string
* **tag**: string OR list of strings
* **time**: string
* **with-atime**: true / false
* **source**: string OR list of strings

`[profile.retention]`

Flags used by resticprofile only

* **before-backup**: true / false
* **after-backup**: true / false
* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-log**: string

Flags passed to the restic command line

* **keep-last**: integer
* **keep-hourly**: integer
* **keep-daily**: integer
* **keep-weekly**: integer
* **keep-monthly**: integer
* **keep-yearly**: integer
* **keep-within**: string
* **keep-tag**: string OR list of strings
* **host**: true / false OR string
* **tag**: string OR list of strings
* **path**: string OR list of strings
* **compact**: true / false
* **group-by**: string
* **dry-run**: true / false
* **prune**: true / false

`[profile.snapshots]`

Flags passed to the restic command line

* **compact**: true / false
* **group-by**: string
* **host**: true / false OR string
* **last**: true / false
* **path**: string OR list of strings
* **tag**: string OR list of strings

`[profile.forget]`

Flags passed to the restic command line

* **keep-last**: integer
* **keep-hourly**: integer
* **keep-daily**: integer
* **keep-weekly**: integer
* **keep-monthly**: integer
* **keep-yearly**: integer
* **keep-within**: string
* **keep-tag**: string OR list of strings
* **host**: true / false OR string
* **tag**: string OR list of strings
* **path**: string OR list of strings
* **compact**: true / false
* **group-by**: string
* **dry-run**: true / false
* **prune**: true / false

`[profile.check]`

Flags used by resticprofile only

* **schedule**: string OR list of strings
* **schedule-permission**: string (`user` or `system`)
* **schedule-log**: string

Flags passed to the restic command line

* **check-unused**: true / false
* **read-data**: true / false
* **read-data-subset**: string
* **with-cache**: true / false

`[profile.mount]`

Flags passed to the restic command line

* **allow-other**: true / false
* **allow-root**: true / false
* **host**: true / false OR string
* **no-default-permissions**: true / false
* **owner-root**: true / false
* **path**: string OR list of strings
* **snapshot-template**: string
* **tag**: string OR list of strings

## Appendix

As an example, here's a similar configuration file in YAML:

```yaml
global:
    default-command: snapshots
    initialize: false
    priority: low

groups:
    full-backup:
    - root
    - src

default:
    env:
        tmp: /tmp
    password-file: key
    repository: /backup

documents:
    backup:
        source: ~/Documents
    repository: ~/backup
    snapshots:
        tag:
        - documents

root:
    backup:
        exclude-caches: true
        exclude-file:
        - root-excludes
        - excludes
        one-file-system: false
        source:
        - /
        tag:
        - test
        - dev
    inherit: default
    initialize: true
    retention:
        after-backup: true
        before-backup: false
        compact: false
        host: true
        keep-daily: 1
        keep-hourly: 1
        keep-last: 3
        keep-monthly: 1
        keep-tag:
        - forever
        keep-weekly: 1
        keep-within: 3h
        keep-yearly: 1
        prune: false
        tag:
        - test
        - dev

self:
    backup:
        source: ./
    repository: ../backup
    snapshots:
        tag:
        - self

src:
    lock: "/tmp/resticprofile-profile-src.lock"
    backup:
        check-before: true
        exclude:
        - /**/.git
        exclude-caches: true
        one-file-system: false
        run-after: echo All Done!
        run-before:
        - echo Starting!
        - ls -al ~/go
        source:
        - ~/go
        tag:
        - test
        - dev
    inherit: default
    initialize: true
    retention:
        after-backup: true
        before-backup: false
        compact: false
        keep-within: 30d
        prune: true
    snapshots:
        tag:
        - test
        - dev
        
stdin:
    backup:
        stdin: true
        stdin-filename: stdin-test
        tag:
        - stdin
    inherit: default
    snapshots:
        tag:
        - stdin

```

Also here's an example of a configuration file in HCL:
```hcl
global {
    priority = "low"
    ionice = true
    ionice-class = 2
    ionice-level = 6
    # don't start if the memory available is < 1000MB
    min-memory = 1000
}

groups {
    all = ["src", "self"]
}

default {
    repository = "/tmp/backup"
    password-file = "key"
    run-before = "echo Profile started!"
    run-after = "echo Profile finished!"
    run-after-fail = "echo An error occurred!"
}


src {
    inherit = "default"
    initialize = true
    lock = "/tmp/backup/resticprofile-profile-src.lock"

    snapshots = {
        tag = [ "test", "dev" ]
    }

    backup = {
        run-before = [ "echo Starting!", "ls -al ~/go/src" ]
        run-after = "echo All Done!"
        exclude = [ "/**/.git" ]
        exclude-caches = true
        tag = [ "test", "dev" ]
        source = [ "~/go/src" ]
        check-before = true
    }

    retention = {
        before-backup = false
        after-backup = true
        keep-last = 3
        compact = false
        prune = true
    }

    check = {
        check-unused = true
        with-cache = false
    }
}

self {
    inherit = "default"
    initialize = false

    snapshots = {
        tag = [ "self" ]
    }

    backup = {
        source = "./"
        tag = [ "self" ]
    }
}

# sending stream through stdin

stdin = {
    inherit = "default"

    snapshots = {
        tag = [ "stdin" ]
    }

    backup = {
        stdin = true
        stdin-filename = "stdin-test"
        tag = [ "stdin" ]
    }
}

```

## Using resticprofile and systemd

systemd is a common service manager in use by many Linux distributions.
resticprofile has the ability to create systemd timer and service files.
systemd can be used in place of cron to schedule backups.

User systemd units are created under the user's systemd profile (~/.config/systemd/user).

System units are created in /etc/systemd/system

### systemd calendars

resticprofile uses systemd
[OnCalendar](https://www.freedesktop.org/software/systemd/man/systemd.time.html#Calendar%20Events)
format to schedule events.

Testing systemd calendars can be done with the systemd-analyze application.
systemd-analyze will display when the next trigger will happen:

```
$ systemd-analyze calendar 'daily'
  Original form: daily
Normalized form: *-*-* 00:00:00
    Next elapse: Sat 2020-04-18 00:00:00 CDT
       (in UTC): Sat 2020-04-18 05:00:00 UTC
       From now: 10h left
```

### First time schedule

When you schedule a profile with the `schedule` command, under the hood resticprofile will
- create the unit file
- create the timer file
- run `systemctl daemon-reload` (only if `schedule-permission` is set to `system`)
- run `systemctl enable`
- run `systemctl start`


## Using resticprofile and launchd on macOS

`launchd` is the service manager on macOS. resticprofile can schedule a profile via a _user agent_ or a _daemon_ in launchd.

### User agent

A user agent is generated when you set `schedule-permission` to `user`.

It consists of a `plist` file in the folder `~/Library/LaunchAgents`:

A user agent **mostly** runs with the privileges of the user. But if you backup some specific files, like your contacts or your calendar for example, you will need to give more permissions to resticprofile **and** restic.

For this to happen, you need to start the agent or daemon from a console window first (resticprofile will ask if you want to do so)

If your profile is a backup profile called `remote`, the command to run manually is:

```
% launchctl start local.resticprofile.remote.backup
```

Once you grant the permission, the background agents/daemon will be able to run normally.

There's some information in this thread: https://github.com/restic/restic/issues/2051

*TODO: I'm going to try to compile a comprehensive how-to guide from all the information from the thread. Stay tuned!*

#### Special case of schedule-permission=user with sudo

Please note if you schedule a user agent while running resticprofile with sudo: the user agent will be registered to the root user, and not your initial user context. It means you can only see it (`status`) and remove it (`unschedule`) via sudo.

### Daemon

A launchd daemon is generated when you set `schedule-permission` to `system`. 

It consists of a `plist` file in the folder `/Library/LaunchDaemons`. You have to run resticprofile with sudo to `schedule`, check the  `status` and `unschedule` the profile.
