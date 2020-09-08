### Notes

### Incremental Improvement

#### Execution profile support

Add execution profile usage. It is to help the seamless integration with CI/CD pipeline

For example, normally, for your CI/CD pipeline, you will need a list of env vars as input of your script/cli command

eg.

```

export ENV_VAR1=var1
export ENV_VAR2=var2
...........
export ENV_VAR10=var10

source ./scripts/aws_access.rc && ./my_cli.sh

```

Now, with UPcmd, you do it this way:

in up.yml, you config an exec profile entry:

```
eprofiles:
  - name: dev_test
    instance: dev
    taskname: CreateMyAppStack
    verbose: vvv

    evars:
      - name: APP_NAME
        value: my_dev_test_app
        
      - name: ENV_VAR10
        value: env_var10        

```

In this case, we simply use one evar: APP_NAME to present as one of a list of env vars as example. 

* instance: this links to the insanceid in the scope, in this case, dev
* taskname: optional
* verbose:  optional

If you choose to ues optional params: taskname and verbose, you simplify the up cli args

Now to execute your pipeline, you simply setup one ENV var in your pipeline:

```
export EProfileID=dev_test
```

then run below entry point cli command:

```
./upngo.sh
```

#### Summary

The exec profile simplifies the pipeline setup by eliminating all environment variable setup either in GUI, or via other mechanisms. Now you only need one ENV var - EProfileID, then just use the entry point script upngo.sh for any task. UPcmd will use the config to pick up right task and use verbose level you prefer. Also, you could just configure your pipeline to be triggered automatically, so that your code push will automatically trigger the pipeline and you don't have to use GUI to click button at all. You can use condition in your workflow to determine how you want to proceed.
 
 
#### demo

Note about the envrc: this is for demo only, in your pipeline, you could use env vars or physical file for the encryption key for strengthened security

```
. ./envrc
export EProfileID=dev_test
./upngo.sh

↑126 0002/04 git:(master) ▶ . ./envrc
export EProfileID=dev_test
./upngo.sh

  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   648  100   648    0     0   1655      0 --:--:-- --:--:-- --:--:--  1653
100 20.9M  100 20.9M    0     0  2649k      0  0:00:08  0:00:08 --:--:-- 4401k
eprofileid: dev_test
loading [Config]:  ./upconfig.yml
Main config:
             Version -> 1.0.0
              RefDir -> ./ups
             WorkDir -> cwd
          AbsWorkDir -> /up-project/up/examples/0002/04
            TaskFile -> up.yml
             Verbose -> v
          ModuleName -> self
           ShellType -> /bin/sh
       MaxCallLayers -> 8
             Timeout -> 3600000
 MaxModuelCallLayers -> 256
           EntryTask -> Main
work dir: /up-project/up/examples/0002/04
-exec task: Main
loading [Task]:  ./up.yml
module: [self], instance id: [dev], exec profile: [dev_test]
profile - dev_test envVars:

(*core.Cache)({
  "envVar_APP_NAME": "my_dev_test_app"
})

loading [./main.yml]:  ./ups/./main.yml
loading [./utils/encrypt.yml]:  ./ups/./utils/encrypt.yml
loading [./utils/venv.yml]:  ./ups/./utils/venv.yml
loading [./myapp/create.yml]:  ./ups/./myapp/create.yml
Task6: [CreateMyAppStack ==> CreateMyAppStack:  ]
-Step1:
self: final context exec vars:

(*core.Cache)({
  "up_runtime_task_layer_number": 0,
  "api_password": "Eu6wFdmnoV4gBFpq6lRq/5HU3ATgXa9BbFjaKrXp/pcD+x4WpT3ot1xC9QBGtzVS",
  "secure_api_password": "the_api_password",
  "api_ep": "http://httpbin.org/post",
  "a": "dev-a",
  "app_name": "my_dev_test_app",
  "api_username": "ixAvykgdH73SafoaGEGB+WiPH/zwZzYQnDMUrIig7lc=",
  "secure_api_username": "api_username",
  "enc_key": "Jb9SVdEy2!!S@WjJ"
})

=Task5: [CreateMyAppStack ==> set_aws_credential:  ]
--Step1:
self: final context exec vars:

(*core.Cache)({
  "api_username": "ixAvykgdH73SafoaGEGB+WiPH/zwZzYQnDMUrIig7lc=",
  "up_runtime_task_layer_number": 1,
  "api_password": "Eu6wFdmnoV4gBFpq6lRq/5HU3ATgXa9BbFjaKrXp/pcD+x4WpT3ot1xC9QBGtzVS",
  "app_name": "my_dev_test_app",
  "a": "dev-a",
  "secure_api_username": "api_username",
  "enc_key": "Jb9SVdEy2!!S@WjJ",
  "secure_api_password": "the_api_password",
  "api_ep": "http://httpbin.org/post"
})

~~SubStep1: [virtualEnv: source all needed aws credential and apply accross the full execution session ]
-sourcing execution result:
start sourcing .....
end sourcing .....

-Step2: [create_my_application: fake mock up only ]
self: final context exec vars:

(*core.Cache)({
  "secure_api_password": "the_api_password",
  "api_username": "ixAvykgdH73SafoaGEGB+WiPH/zwZzYQnDMUrIig7lc=",
  "secure_api_username": "api_username",
  "api_ep": "http://httpbin.org/post",
  "enc_key": "Jb9SVdEy2!!S@WjJ",
  "up_runtime_task_layer_number": 1,
  "a": "dev-a",
  "app_name": "my_dev_test_app",
  "api_password": "Eu6wFdmnoV4gBFpq6lRq/5HU3ATgXa9BbFjaKrXp/pcD+x4WpT3ot1xC9QBGtzVS"
})

cmd( 1):
echo "AWS_SESSION_TOKEN is - ${AWS_SESSION_TOKEN}"

-
AWS_SESSION_TOKEN is - my_aws_session_token

-
 .. ok
cmd( 2):
curl -s -d """
{
  "name": "tom",
  "class": "year12-k",
  "school": "SG"
  "username": "{{.secure_api_username}}"
  "password": "{{.secure_api_password}}"
  "token": "${AWS_SESSION_TOKEN}"
}""" \
-X POST \
-H "accept: application/json" \
{{.api_ep}}


-
{
  "args": {}, 
  "data": "", 
  "files": {}, 
  "form": {
    "\n{\n  name: tom,\n  class: year12-k,\n  school: SG\n  username: api_username\n  password: the_api_password\n  token: my_aws_session_token\n}": ""
  }, 
  "headers": {
    "Accept": "application/json", 
    "Content-Length": "133", 
    "Content-Type": "application/x-www-form-urlencoded", 
    "Host": "httpbin.org", 
    "User-Agent": "curl/7.54.0", 
    "X-Amzn-Trace-Id": "Root=1-5f579add-2039c4c80b917e9004ab1ac8"
  }, 
  "json": null, 
  "origin": "118.211.180.66", 
  "url": "http://httpbin.org/post"
}

-
 .. ok
. ok

```