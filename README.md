# UPcmd [docs site](https://upcmd.netlify.app/)

The Ultimate Provisioner: the modern configuration management, build and automation tool

## [UPcmd  - The Ultimate Provisioner](https://upcmd.netlify.app/usage/cli_usage/)

UP is designed and implemented to shine as a modern tool for below:

  * Configuration management
  * Build, continuous delivery, integration with CI/CD
  * Comprehensive workflow orchestration: full support of almost all type of condition, loop(recursive), break
  * Flexible configuration organisation
  * No dependency hell issue
  * Precise modeling, data is object, object is the data
  * Design of composition, separate func type, data and implementation
  * Use inteface(call func) for abstraction of intention, data input and implementation
  * Many builtin dry run, assert, debugging features for the insights, developer friendly
  * ... many more for you to discover, check out the docs

It is a build tool like Ansible, Make, Rake, Ant, Gradle, Puppet, Taskfile etc, but it is a little smarter to try to make things a little easier.

The goal of UP is to provide a quick (I'd say the quickest) solution to enable continuous integration and continuous deployment (CI/CD). It is simple to use and yet powerful to address many common challenges nowadays devop teams facing in the Cloud environment.

It is designed with mindful consideration of collaboration with automation in Kubernetes, helm charts, api call.

It follows best practices integrating with common CI/CD tools, such as GOCD, Jenkins, Drone, Gitlab CI and so on and is a good company of all types of CLI tools.

It brings a fun DSL programming interface, a way of modeling and engineering into CLI. It enables OO design and rapid Test Driven Development (TDD) and shorter software delivery cycle.

* Hello, world

```
tasks:
  -
    name: task
    task:
      -
        func: shell
        do:
          - echo "hello, world"
```

* Appetizers

Below shows:

* It is a simple deployment of a web application, it has a prior step of database deployment
* The database deployment will depend on the db configuration base on per environment
* the non prod envs: staging and dev share the same database configurations using a shared database 
* the non prod envs: staging and dev each individually will use different database host name though
* a dvar db_password using aes to manage the password   

To deploy, simply specify a instanceid to be associated with an environment, eg: dev, staging or prod

```
up ngo -i staging
```  

The config: 
``` 

scopes:
- name: global
  vars:
    db_driver: postgres
    port: 5432

- name: nonprod
  members:
  - dev
  - staging
  vars:
    db_host: nonpord_database.test.host
    db_user: test_db_user
    db_password: could_be_encrypted_using_upcmd_too
  dvars:
    - name: db_password
      value: '6HmsmiJIW1PfIXcF4WwOKOMDiL7PstgfKs2aRFajrwY='

- name: prod
  members: [prod]
  vars:
    host_alias: prod

- name: dev
  vars:
    host_alias: dev

- name: staging
  vars:
    host_alias: staging

- name: prod
  vars:
    host_alias: prod
    db_host: pord_database.proddb.host
    db_user: prod_db_user
  dvars:
    - name: db_password
      value: 'prod_encrypte_aes'

dvars:
  - name: db_hostname
    value: '{{.host_alias}}.myapp.com'
  - name: db_url
    value: 'jdbc:{{.db_driver}}://{{.db_hostname}}:{{.db_port}}/test?user={{.db_user}}&password={{.db_password}}&ssl=true'

tasks:
  -
    name: Main
    desc: deploy my web app stack
    task:
      -
        func: call
        do:
          - deploy_database
          - deploy_web

      -
        func: shell
        do:
          - systemctl start my_database 
          - systemctl start my_web_server

  -
    name: deploy_web
    task:
      -
        func: shell
        do:
          - deploy myweb_server

  -
    name: deploy_database
    task:
      -
        func: shell
        do:
          - deploy mydatabase

```

With the evolving of the up.yml file, you could externalize the configuration to individual files or make them as module to be reused or shared. Please check out the doc for more details.

### High level design

At high level, UPcmd process flows like below:

* The process engine process the scope vars and merge them with global vars, then in the run time it will merge with local vars again
* For the callee task, the local vars will be overriden by the vars passed from caller task
    
![high level design](https://raw.githubusercontent.com/upcmd/updocs/master/static/up_high_level.png)

### Possible applications

UPcmd is a generic automation tool, given your automation solution being backed by Unix Shell. You do not need Shell executable though, as it has default GOSH builtin just in case you will need one to fall back to.

There could be application as below, but not limit to: 
* Build, package, publish, test, deploy for all different types of applications in your local machine, or integrate with CI/CD tools/pipelines
* UPcmd could be used as tool/platform/pipeline agnostic abstraction layer, leave the most configuration to UPcmd to manage as an execution profile, expose only the profileid to be linked with the tool, eg: jenkins/gitlab ci, so that all your automation is portable. In case you need to switch from one to another, you don't need to rewrite all the automation. In this case, UPcmd's configured tasks could be regarded as pipeline as code.  
* A collection of util like (tool box) for local machine automation, for example, 
    1. bootstrap the whole Macbook with all upgraded packages, setup all your dotfiles
    2. bootstrap the whole Linux box/virtual VM/vagrant box/docker container
* Create CLI program, prompt with user input, encrypt/decrypt secrets
* Web service/rest api call and message transformation
* Cloud service provisioning, eg drive complicated workflow to manage to create full application stack in AWS or k8s cluster, utilise and integrate with other CLI commands, such as packer, aws cli, kubectl, helm, terraform
* Reuse/consume or share modules to deal with a particular use case. 
* Resolve the dependencies issue by simply invoking different version of the relevant CLI/docker run
* The orchestration of UPcmd task itself could be seen as prototyping tool and design tool, or use the defined workflow as skeleton to guide the implementation from different part        

### Quick install

#### Install the binary

There are 32 different distro for different combination of OS and Arch type, check them out: [release](https://github.com/upcmd/up/releases)

Below are common one for latest tagged stable release:

* Install for Mac OSX:

```
curl -s https://api.github.com/repos/upcmd/up/releases \
    |grep darwin_amd64_v \
    |grep download \
    |head -n 1 \
    |awk '{print $2}' \
    |xargs -I % curl -L % -o up \
    && chmod +x up
```

* Install for Linux:

```
curl -s https://api.github.com/repos/upcmd/up/releases \
    |grep linux_amd64_v \
    |grep download \
    |head -n 1 \
    |awk '{print $2}' \
    |xargs -I % curl -L % -o up \
    && chmod +x up
```

* Install for Windows:

```
curl -s https://api.github.com/repos/upcmd/up/releases \
    |grep windows_amd64_v \
    |grep download \
    |head -n 1 \
    |awk '{print $2}' \
    |xargs -I % curl -L % -o up \
    && chmod +x up
```

Below are for latest bleeding-edge release from master branch:

* Install for Mac OSX:

```
curl -s https://api.github.com/repos/upcmd/up/releases \
    |grep darwin_amd64_latest \
    |grep download \
    |head -n 1 \
    |awk '{print $2}' \
    |xargs -I % curl -L % -o up \
    && chmod +x up
```

* Install for Linux:

```
curl -s https://api.github.com/repos/upcmd/up/releases \
    |grep linux_amd64_latest \
    |grep download \
    |head -n 1 \
    |awk '{print $2}' \
    |xargs -I % curl -L % -o up \
    && chmod +x up
```

* Install for Windows:

```
curl -s https://api.github.com/repos/upcmd/up/releases \
    |grep windows_amd64_latest \
    |grep download \
    |head -n 1 \
    |awk '{print $2}' \
    |xargs -I % curl -L % -o up \
    && chmod +x up
```

Move the downloaded UP command executable to an executable path, eg. /usr/local/bin, so you can use it system wide

#### Install from source

Ensure you use go 1.14 (prefered)

```
go install github.com/upcmd/up/app/up
```

The up CLI command will be installed to: $HOME/go/bin, make sure you have this in your PATH

#### Usage: followup the detailed documentation site: [https://upcmd.netlify.app/](https://upcmd.netlify.app/usage/cli_usage/)

### A little taste of UPcmd

Below is a simple greeting example, and a list, inspect and execution view of the task.

* With some smarts: logic and loop etc [doc](https://upcmd.netlify.app/quick-start/c0151/) | [source](https://github.com/upcmd/up/blob/master/tests/functests/c0151.yml)

This shows:
* the greet task is an implementation, by default it was called with default global var greet_to value, but with supply of local var of "Grace", it changes the behaviour [see concept of interface](https://upcmd.netlify.app/call-func/c0020/)
* loop through
* if/else logic 
* chain through tasks

```

Vars:
  greet_to: Tom
  weather: sunny

tasks:
  -
    name: task
    desc: main task of hello world demo of UPcmd
    task:
      -
        func: call
        desc: greet to Tom
        do:
          - greet

      -
        func: call
        desc: greet to Grace
        vars:
          greet_to: Grace
        do:
          - greet


      -
        func: cmd
        desc: do  you get the idea?
        do:
          - name: print
            cmd: |
              Have you got a little taste of using the UPcmd?

      -
        func: call
        desc: greet to a team
        vars:
          team:
            - Jason
            - Connie
          weather: stormy
        loop: team
        do:
          - sayhi

  -
    name: greet
    desc: greet to some one
    task:
      -
        func: shell
        desc: say hello
        do:
          - echo "Hello, {{.greet_to}}"

      -
        func: cmd
        desc: talk about weather
        do:
          - name: print
            cmd: 'It is {{.weather}}'

      -
        func: cmd
        desc: ice break
        do:
          - name: print
            cmd: 'What a great day!'
        if: '{{eq .weather "sunny"}}'
        else:
          -
            func: cmd
            do:
              - name: print
                cmd: 'What a bad day!!'

  -
    name: sayhi
    desc: say hi to some one
    task:
      -
        func: cmd
        desc: say hi to someone
        do:
          - name: print
            cmd: 'Hi {{.loopitem}}, how are you?'

      -
        func: call
        desc: greet to the team member
        dvars:
          - name: greet_to
            value: '{{.loopitem}}'
        do:
          - greet

```

![A little taste](https://raw.githubusercontent.com/upcmd/updocs/master/static/a_little_taste.png)

### Demo

It demos:
* create upcmd task skeleton using init command
* show the intro demo code and execution
* use module
* test driven, assert and color print

Check it out yourself: [source](https://github.com/upcmd/up-demo/blob/master/demo.sh) and try to have fun to run though the examples by yourself

![demo](https://raw.githubusercontent.com/upcmd/up-demo/master/intro.gif)

###  Why yet another build tool

* Make was initially designed and used for building C program, even though it could be adopted for other purpose, some of the hard to learn trivials often cause problems than the benefits added to the team, and it is burning the brain. It is hard to make automation task extended to a more advanced level, readbility degrades rapidly and it is risky to implement critical logic using Make. Make is just a little old for modern business requirements. (Sorry, maybe this is just from some one not good at using Makefile)

* Rake is smart and powerful. If you don't mind learning a little bit of Ruby, it is a good choice of building tool. Similarly to Ant, Gradle. They are all bind to a language specific, it is just not right when it comes to the case that you want to automate things in cloud environment. In most of cases when it requires automation in a cloud environment, in a given spun up AWS EC2 instance, a shell session, a kubernete pod, you would want some thing just works without any dependencies. You simply do not want to mantain the consistency of chain of upgrding path for all language pkgs in multiple environments. In these cases, Rake, Gradle, Ant are not best options.

* Ansible, Puppet are configuration management tools. They are powerful, there are many builtin well tested modules you could use. However Ansible might be too huge for little job and most of the time it tends to over kill, also it suffers the same problem of python/python packages dependencies.

  A common usage of Ansible for many teams is to use the local ssh execution with group/host vars for templating and workflow automation, which is simply not right. Also the way the vars being managed is not fine grained. The ansible role as a reusable module is not flexible to implement more complicated tasks.

* Inspired by https://taskfile.dev/,  it is tiny tool making build and automation easier and elegant, however it lacks some of the features in a practical cloud environment for CI/CD, devops automation, hence this project is born for that purpose

### Features

1. Drop in replacement for Makefile, but way more powerful. It uses a composition model rather than dependency model for flexibility/composibility
2. Implemented in golang, so no dependency hell, no maintenance of runtime and ensure the version consistency across multiple/many execution contexts
3. Use scopes to manage group of execution context, the variables associated with the scope. Fine grained scoping model to support variable auto overriding/merging. Similar to Ansible global/group vars, host vars, but more powerful to support leaf level objects auto merging
4. Use dvar - dynamic var, a special design to achieve many incredible features, for example:
    * manage security: encrypt/decrypt, like ansible-vault, builtin
    * dynamics on dynamics: it allows you to specify how many layers of expansion you'd like to dynamically render a variable
    * builtin templating capability
    * use golang template, supporting all(220+) (builtin|sprig|gtf) funcs/pipeline so that your configuration could be well controlled in template using objects
    * auto message transformation between yaml|json result to object used internally
    * conform the hierarchical scoping model for var merging to leaf level
    * manage setup/read env vars in the same scoping model so that you could have seamless integration with minimal exposed demanding ENV vars from CD/CI tool
    * auto validation for mandatory vars
5. Color print and adjustable verbose level
6. Flexible programming model to allow you to separate implementation with interface so that common code could be reused via task_ref
Allow empty skeleton to be laid for testing driving process or guide as seudo code, but fill in the details and implementation gradually
7. Flow control:
      * ignore_error
      * dry run
      * if condition support
      * loop support to iterate through a list of items
      * mult-layered loop and break
      * block and embedded block of code for execution
8. Flexible configuration style to load dvar, scope, flow from external yaml so that the programming code will be a little cleaner and organised. Your code could evolve starting from simple, then externalize detailed implementation to files.
9. Support the unlimited yml object, so yml config in var is text and it is also object.It could be merged in scopes automatically, it could be processed using go template
10. Battery included for common builtin commands: print, reg, dereg, template, readfile, writefile
11. Builtin yml liter and object query, modification
12. Call func is really shining powerful design to be used:
    * Compose the sequential execution of block of code
    * Use the stack design, so it segregates all its local vars so that the vars used in its implementation will not pollute the caller's vars
    * It serves like a interface to separates the goal and implementation and makes the code is reusable
13. The shell execution binary is configurable, builtin support for GOSH (mvdan.cc/sh). This means that you do not need native shell/bash/zsh installed in order for task execution, you can run task from a windows machine.
14. It provides a module mechanism to encourage community to share modular code so that you do not need to reinvent the wheel to develop the same function again

### Real Examples

Both UPcmd project build and the docs entire site build use the UPcmd itself

##### Project release for UPcmd [source](https://github.com/upcmd/up/blob/master/up.yml)
```
up ngo publish
```

##### Documentation [doc site](https://upcmd.netlify.app/)

build of the entire doc site using one build task: 
* [source](https://github.com/upcmd/updocs/blob/master/up.yml)
* [details](https://upcmd.netlify.app/advanced-cases/upcmd-doc-gen/)

##### A web scripting example [how?](https://upcmd.netlify.app/advanced-cases/web-scraping/)

### Testing

There are over 200~ test cases, every release come with a full passed regression test of all cases defined, [source](https://github.com/upcmd/up/tree/master/tests)

* [common examples](https://github.com/upcmd/up/tree/master/tests/functests)
* [module usage examples](https://github.com/upcmd/up/tree/master/tests/modtests)

These test cases are not only about the tests, they are the usage examples with documentation self explaned

### License

This project is under MPL-2.0 License
