package main

import (
    "fmt"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ec2"
    "github.com/rivo/tview"
)

type InstanceInfo struct {
    Id    string
    Name  string
    State string
}

func createSession() (*ec2.EC2, error) {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("eu-west-1"), // Change this according to your region
    })
    if err != nil {
        return nil, err
    }
    return ec2.New(sess), nil
}

func getInstancesInfo(svc *ec2.EC2, instanceID string) ([]InstanceInfo, error) {
    input := &ec2.DescribeInstancesInput{
        InstanceIds: []*string{aws.String(instanceID)}, // Filter for the specific instance
    }
    result, err := svc.DescribeInstances(input)
    if err != nil {
        return nil, err
    }

    var instances []InstanceInfo
    for _, reservation := range result.Reservations {
        for _, instance := range reservation.Instances {
            name := "No Name" // Default name if no Name tag is found
            for _, tag := range instance.Tags {
                if *tag.Key == "Name" {
                    name = *tag.Value
                    break
                }
            }
            instances = append(instances, InstanceInfo{
                Id:    *instance.InstanceId,
                Name:  name,
                State: *instance.State.Name,
            })
        }
    }
    return instances, nil
}

func startInstance(svc *ec2.EC2, instanceId string) error {
    input := &ec2.StartInstancesInput{
        InstanceIds: []*string{aws.String(instanceId)},
    }
    _, err := svc.StartInstances(input)
    return err
}

func stopInstance(svc *ec2.EC2, instanceId string) error {
    input := &ec2.StopInstancesInput{
        InstanceIds: []*string{aws.String(instanceId)},
    }
    _, err := svc.StopInstances(input)
    return err
}

func showInstancesList(app *tview.Application, svc *ec2.EC2, instanceID string) {
    instancesInfo, err := getInstancesInfo(svc, instanceID)
    if err != nil {
        fmt.Println("Error retrieving EC2 instance information:", err)
        return
    }

    list := tview.NewList()
    for _, info := range instancesInfo {
        var stateColor string
        switch info.State {
        case "running":
            stateColor = "[green]"
        case "stopped":
            stateColor = "[red]"
        case "stopping":
            stateColor = "[yellow]"
        case "pending":
            stateColor = "[brown]"
        default:
            stateColor = "[white]"
        }

        displayText := fmt.Sprintf("%s (%s) - %s%s[white]", info.Name, info.Id, stateColor, info.State)
        list.AddItem(displayText, "", 0, makeInstanceSelectionHandler(app, svc, info))
    }

    list.SetBorder(true).SetTitle("Select an instance").SetBorderPadding(1, 1, 1, 1)
    app.SetRoot(list, true).SetFocus(list)
}

func makeInstanceSelectionHandler(app *tview.Application, svc *ec2.EC2, info InstanceInfo) func() {
    return func() {
        actionModal := tview.NewModal().
            SetText(fmt.Sprintf("Do you want to start or stop %s?", info.Id)).
            AddButtons([]string{"Start", "Stop", "Cancel"}).
            SetDoneFunc(func(buttonIndex int, buttonLabel string) {
                switch buttonLabel {
                case "Start":
                    err := startInstance(svc, info.Id)
                    if err != nil {
                        fmt.Println("Error starting the instance:", err)
                    }
                    showInstancesList(app, svc, info.Id) // Reload the instance list
                case "Stop":
                    err := stopInstance(svc, info.Id)
                    if err != nil {
                        fmt.Println("Error stopping the instance:", err)
                    }
                    showInstancesList(app, svc, info.Id) // Reload the instance list
                case "Cancel":
                    showInstancesList(app, svc, info.Id) // Return to the list
                }
            })
        app.SetRoot(actionModal, false).SetFocus(actionModal)
    }
}

func main() {
    svc, err := createSession()
    if err != nil {
        fmt.Println("Error creating AWS session:", err)
        return
    }

    app := tview.NewApplication()
    instanceID := "i-xxxx" // Your specific instance ID

    showInstancesList(app, svc, instanceID)

    if err := app.Run(); err != nil {
        fmt.Printf("Error running the application: %v\n", err)
    }
}
