### Notes

Demo the basic structure to config scope and execute a mock call to create an app stack by send a post api call

### Incremental Improvement

#### Add an interactive encryption/decryption util

```

Ξ 0002/02 git:(master) ▶ up ngo Utils_crypt_interactive -i dev
loading [Config]:  ./upconfig.yml
Main config:
             Version -> 1.0.0
              RefDir -> ./ups
             WorkDir -> cwd
          AbsWorkDir -> /up-project/up/examples/0002/02
            TaskFile -> up.yml
             Verbose -> v
          ModuleName -> self
           ShellType -> /bin/sh
       MaxCallLayers -> 8
             Timeout -> 3600000
 MaxModuelCallLayers -> 256
           EntryTask -> Utils_crypt_interactive
work dir: /up-project/up/examples/0002/02
-exec task: Utils_crypt_interactive
loading [Task]:  ./up.yml
module: [self], instance id: [dev], exec profile: []
loading [./main.yml]:  ./ups/./main.yml
loading [./utils/encrypt.yml]:  ./ups/./utils/encrypt.yml
Task2: [Utils_crypt_interactive ==> Utils_crypt_interactive:  ]
-Step1:
Enter Value For [choice]: 
choose 1 to encrypt or anyting else to decrypt
1
=Task3: [Utils_crypt_interactive ==> encrypt:  ]
--Step1:
Enter Value For [raw]: 
This will be saved as raw's value
api_username
~~SubStep1: [print:  ]
api_username
~~SubStep2: [print:  ]
ixAvykgdH73SafoaGEGB+WiPH/zwZzYQnDMUrIig7lc=
~~SubStep3: [print:  ]
api_username

```

#### Move core business task to sit under ./ups/myapp/create.yml

#### Move the Main task to: ./ups/main.yml

#### Use secret username/password instead of plain text in post call

#### Save the encrypted config vars in up.yml in its own scope

So, you can have one common password for prod/nonprod, or dev/staging individually for nonprod env. Just put them into the scope


### list

```
Ξ 0002/02 git:(master) ▶ up list                       
loading [Config]:  ./upconfig.yml
Main config:
             Version -> 1.0.0
              RefDir -> ./ups
             WorkDir -> cwd
          AbsWorkDir -> /up-project/up/examples/0002/02
            TaskFile -> up.yml
             Verbose -> v
          ModuleName -> self
           ShellType -> /bin/sh
       MaxCallLayers -> 8
             Timeout -> 3600000
 MaxModuelCallLayers -> 256
           EntryTask -> 
work dir: /up-project/up/examples/0002/02
loading [Task]:  ./up.yml
module: [self], instance id: [nonamed], exec profile: []
 WARN: [*be aware*] - [both instance id and exec profile are not set]
loading [./main.yml]:  ./ups/./main.yml
loading [./utils/encrypt.yml]:  ./ups/./utils/encrypt.yml
loading [./myapp/create.yml]:  ./ups/./myapp/create.yml
-task list
     1  |                    Main |   public|  
     2  | Utils_crypt_interactive |   public|  
     3  |                 encrypt |protected|  
     4  |                 decrypt |protected|  
     5  |        CreateMyAppStack |   public|  
```

#### exec

```

Ξ 0002/02 git:(master) ▶ up ngo CreateMyAppStack -i dev
loading [Config]:  ./upconfig.yml
Main config:
             Version -> 1.0.0
              RefDir -> ./ups
             WorkDir -> cwd
          AbsWorkDir -> /up-project/up/examples/0002/02
            TaskFile -> up.yml
             Verbose -> v
          ModuleName -> self
           ShellType -> /bin/sh
       MaxCallLayers -> 8
             Timeout -> 3600000
 MaxModuelCallLayers -> 256
           EntryTask -> CreateMyAppStack
work dir: /up-project/up/examples/0002/02
-exec task: CreateMyAppStack
loading [Task]:  ./up.yml
module: [self], instance id: [dev], exec profile: []
loading [./main.yml]:  ./ups/./main.yml
loading [./utils/encrypt.yml]:  ./ups/./utils/encrypt.yml
loading [./myapp/create.yml]:  ./ups/./myapp/create.yml
Task5: [CreateMyAppStack ==> CreateMyAppStack:  ]
-Step1: [create_my_application: fake mock up only ]
cmd( 1):
-
{
  "args": {}, 
  "data": "", 
  "files": {}, 
  "form": {
    "\n{\n  \"name\": \"tom\",\n  \"class\": \"year12-k\",\n  \"school\": \"SG\"\n  \"username\": \"api_username\"\n  \"password\": \"the_api_password\"\n}": ""
  }, 
  "headers": {
    "Accept": "application/json", 
    "Content-Length": "123", 
    "Content-Type": "application/x-www-form-urlencoded", 
    "Host": "httpbin.org", 
    "User-Agent": "curl/7.54.0", 
    "X-Amzn-Trace-Id": "Root=1-5f578ccc-a88a77be8fb07696fb993f3e"
  }, 
  "json": null, 
  "origin": "118.211.180.66", 
  "url": "http://httpbin.org/post"
}

-
 .. ok
. ok

```
