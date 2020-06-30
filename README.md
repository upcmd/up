# UPcmd [docs site](https://upcmd.netlify.app/)

The Ultimate Provisioner: the modern configuration management, build and automation tool

## UPcmd  - The Ultimate Provisioner

UP is designed and implemented to address some of the common problems of:

  * configuration management
  * build, continuously delivery, integration with ci/cd

It is a build tool like Ansible, Make, Rake, Ant, Gradle, Puppet, Taskfile etc, but it is a little smarter to try to make things a easier

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

###  Why yet another build tool

* Make was initially designed and used for building C program, even though it could be adopted for other purpose, some of the hard to learn trivials often cause problems than the benefits added to the team, and it is burning the brain. It is hard to make automation task extended to a more advanced level, readbility degrades rapidly and it is risky to implement critical logic using Make. Make is just a little old for modern business requirements. (Sorry, maybe this is just from some one not good at using Makefile)

* Rake is smart and powerful. If you don't mind learning a little bit of Ruby, it is a good choice of building tool. Similarly to Ant, Gradle. They are all bind to a language specific, it is just not right when it comes to the case that you want to automate things in cloud environment. In most of cases when it requires automation in a cloud environment, in a given spun up AWS EC2 instance, a shell session, a kubernete pod, you would want some thing just works without any dependencies. You simply do not want to mantain the consistency of chain of upgrding path for all language pkgs in multiple environments. In these cases, Rake, Gradle, Ant are not best options.

* Ansible, Puppet are configuration management tools. They are powerful, there are many builtin well tested modules you could use. However Ansible might be too huge for little job and most of the time it tends to over kill, also it suffers the same problem of python/python packages dependencies.

  A common usage of Ansible for many teams is to use the local ssh execution with group/host vars for templating and workflow automation, which is simply not right. Also the way the vars being managed is not fine grained. The ansible role as a reusable module is not flexible to implement more complicated tasks.

* Inspired by https://taskfile.dev/,  it is tiny tool making build and automation easier and elegant, however it lacks some of the features in a practical cloud environment for CI/CD, devops automation, hence this project is born for that purpose


### Goal

The goal of UP is to provide a quick (I'd say the quickest) solution to enable continuously integration and continuously deployment (CI/CD). It is simple to use and yet powerful to achieve many common challenges nowadays devops face in the Cloud environment

It is designed with mindful consideration of collaboration with automation in Kubernetes, helm charts, api call

It is also put best practice of integration with common CI/CD tools, such as GOCD, Jenkins, Drone, Gitlab CI and so on

It is bringing a DSL programming into CLI and enable OO design and fast test driven development and delivery cycle

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

#### Project Build [source](https://github.com/upcmd/up/blob/master/up.yml)

#### Documentation [doc site](https://upcmd.netlify.app/)

build of the entire doc site using one build task: [source](https://github.com/upcmd/updocs/blob/master/up.yml)

```
up ngo build

```

### Testing

There are around 200~ test cases tested, [source](https://github.com/upcmd/up/tree/master/tests)

* [common examples](https://github.com/upcmd/up/tree/master/tests/functests)
* [module usage examples](https://github.com/upcmd/up/tree/master/tests/modtests)

These test cases are not only about the tests, they are the usage examples with documentation self explaned


### A little taste of UPcmd

Below shows a simple greeting example, also shows list, inspect and execution of the task

![A little taste](https://raw.githubusercontent.com/upcmd/updocs/master/static/a_little_taste.png)


### Demo

![demo](https://raw.githubusercontent.com/upcmd/up-demo/master/intro.gif)

### License

This project is under MPL-2.0 License
