### Notes

### Incremental Improvement

Add virtualEnv support to preload/source environment variables, this will apply to the entire shell execution context. Checkout the virtualEnv for more details

```
Ξ 0002/03 git:(master) ▶ . ./envrc
Ξ 0002/03 git:(master) ▶ up ngo CreateMyAppStack -i dev
loading [Config]:  ./upconfig.yml
Main config:
             Version -> 1.0.0
              RefDir -> ./ups
             WorkDir -> cwd
          AbsWorkDir -> /up-project/up/examples/0002/03
            TaskFile -> up.yml
             Verbose -> v
          ModuleName -> self
           ShellType -> /bin/sh
       MaxCallLayers -> 8
             Timeout -> 3600000
 MaxModuelCallLayers -> 256
           EntryTask -> CreateMyAppStack
work dir: /up-project/up/examples/0002/03
-exec task: CreateMyAppStack
loading [Task]:  ./up.yml
module: [self], instance id: [dev], exec profile: []
loading [./main.yml]:  ./ups/./main.yml
loading [./utils/encrypt.yml]:  ./ups/./utils/encrypt.yml
loading [./utils/venv.yml]:  ./ups/./utils/venv.yml
loading [./myapp/create.yml]:  ./ups/./myapp/create.yml
Task6: [CreateMyAppStack ==> CreateMyAppStack:  ]
-Step1:
=Task5: [CreateMyAppStack ==> set_aws_credential:  ]
--Step1:
~~SubStep1: [virtualEnv: source all needed aws credential and apply accross the full execution session ]
-sourcing execution result:
start sourcing .....
end sourcing .....

-Step2: [create_my_application: fake mock up only ]
cmd( 1):
-
AWS_SESSION_TOKEN is - my_aws_session_token

-
 .. ok
cmd( 2):
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
    "X-Amzn-Trace-Id": "Root=1-5f5791c8-7de7fc0b046fb0854a53121f"
  }, 
  "json": null, 
  "origin": "118.211.180.66", 
  "url": "http://httpbin.org/post"
}

-
 .. ok
. ok
```