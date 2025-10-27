# AWS EC2 Instance Controller

This application provides a terminal-based interface for managing AWS EC2 instances, allowing users to start, stop, and view the status of their EC2 instances directly from the command line.

## Prerequisites

Before compiling and running this application, ensure the following are installed on your Windows system:

- **Go (Golang):** The application is written in Go. It must be installed to compile the source code. [Download Go](https://golang.org/dl/).

- **AWS CLI:** Configure the AWS Command Line Interface (CLI) with your AWS credentials (Access Key ID and Secret Access Key). Follow the [AWS CLI configuration guide](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html).

You need to create 2 policies , one to show instances.

Policies
Describe

```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "ec2:DescribeInstances",
            "Resource": "*"
        }
    ]
}
```

and one for start stop
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:StartInstances",
                "ec2:StopInstances",
                "ec2:DescribeInstances"
            ],
            "Resource": "arn:aws:ec2:your_aws_region:your_aws_account_id:instance/i-xxxidoftheec2instance"
        }
    ]
}
```

Create the user and attribute him this two policies.

## Download the Source Code

1. **Download the Source Code:** Download the ZIP file from [here](https://github.com/p2zbar/ec2control/archive/refs/heads/main.zip).

2. **Extract the ZIP File:** Extract the contents to a folder on your computer.

## Compilation Instructions

Open a Command Prompt or PowerShell window and navigate to your project directory:

```
cd path\to\folder\of\extracted\main
```

Initialize a Go module (if not already present):

```
go mod init main
```

Fetch and install dependencies:  

```
go mod tidy
```

```
Modify in the main.go in the function main L138 i-xxxx by your real EC2 Instance ID
```


Compile the application to create an executable:  
```  
go build -o ec2controller.exe main.go
```

Run your ec2controller.exe  

https://github.com/p2zbar/ec2control/assets/125798712/db92f341-5f70-43bd-8407-2b26ebb385ec




