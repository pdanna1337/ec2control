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

func getInstancesInfo(svc *ec2.EC2) ([]InstanceInfo, error) {
    input := &ec2.DescribeInstancesInput{}
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

func showInstancesList(app *tview.Application, svc *ec2.EC2) {
    instancesInfo, err := getInstancesInfo(svc)
    if err != nil {
        fmt.Println("Error retrieving EC2 instance information:", err)
        return
    }

    list := tview.NewList()
    for _, info := range instancesInfo {
        displayText := fmt.Sprintf("%s (%s) - %s", info.Name, info.Id, info.State)
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
                case "Start", "Stop":
                    confirmText := fmt.Sprintf("Are you sure you want to %s %s?", buttonLabel, info.Id)
                    confirmationModal := tview.NewModal().
                        SetText(confirmText).
                        AddButtons([]string{"Confirm", "Cancel"}).
                        SetDoneFunc(func(confirmIndex int, confirmLabel string) {
                            if confirmLabel == "Confirm" {
                                if buttonLabel == "Start" {
                                    err := startInstance(svc, info.Id)
                                    if err != nil {
                                        fmt.Println("Error starting the instance:", err)
                                    }
                                } else {
                                    err := stopInstance(svc, info.Id)
                                    if err != nil {
                                        fmt.Println("Error stopping the instance:", err)
                                    }
                                }
                                showInstancesList(app, svc) // Reload the instance list
                            } else {
                                showInstancesList(app, svc) // Return to the list if "Cancel" is selected
                            }
                        })
                    app.SetRoot(confirmationModal, false).SetFocus(confirmationModal)
                default:
                    showInstancesList(app, svc) // Return to the list if "Cancel" is selected in the first modal
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
    showInstancesList(app, svc)

    if err := app.Run(); err != nil {
        fmt.Printf("Error running the application: %v\n", err)
    }
}