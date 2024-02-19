package main

import (
    "fmt"
    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ec2"
    "github.com/rivo/tview"
)

// Structure pour stocker les informations des instances
type InstanceInfo struct {
    Id    string
    Name  string
    State string
}

// Créer une session AWS
func createSession() (*ec2.EC2, error) {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("eu-west-1"), // Changez cela selon votre région
    })
    if err != nil {
        return nil, err
    }
    return ec2.New(sess), nil
}

// Récupérer les informations des instances EC2
func getInstancesInfo(svc *ec2.EC2) ([]InstanceInfo, error) {
    input := &ec2.DescribeInstancesInput{}
    result, err := svc.DescribeInstances(input)
    if err != nil {
        return nil, err
    }

    var instances []InstanceInfo
    for _, reservation := range result.Reservations {
        for _, instance := range reservation.Instances {
            name := "No Name" // Nom par défaut si aucune balise Name n'est trouvée
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

// Fonction pour démarrer une instance
func startInstance(svc *ec2.EC2, instanceId string) error {
    input := &ec2.StartInstancesInput{
        InstanceIds: []*string{aws.String(instanceId)},
    }
    _, err := svc.StartInstances(input)
    return err
}

// Fonction pour arrêter une instance
func stopInstance(svc *ec2.EC2, instanceId string) error {
    input := &ec2.StopInstancesInput{
        InstanceIds: []*string{aws.String(instanceId)},
    }
    _, err := svc.StopInstances(input)
    return err
}

func main() {
    svc, err := createSession()
    if err != nil {
        fmt.Println("Erreur lors de la création de la session AWS:", err)
        return
    }

    instancesInfo, err := getInstancesInfo(svc)
    if err != nil {
        fmt.Println("Erreur lors de la récupération des informations des instances EC2:", err)
        return
    }

    app := tview.NewApplication()
    list := tview.NewList()
    for _, info := range instancesInfo {
        info := info // Capture la variable dans la portée de la closure
        displayText := fmt.Sprintf("%s (%s) - %s", info.Name, info.Id, info.State)
        list.AddItem(displayText, "", 0, func(info InstanceInfo) func() {
            return func() {
                actionModal := tview.NewModal().
                    SetText(fmt.Sprintf("Voulez-vous démarrer ou arrêter %s?", info.Id)).
                    AddButtons([]string{"Démarrer", "Arrêter", "Annuler"}).
                    SetDoneFunc(func(buttonIndex int, buttonLabel string) {
                        switch buttonLabel {
                        case "Démarrer", "Arrêter":
                            confirmText := fmt.Sprintf("Êtes-vous sûr de vouloir %s %s?", buttonLabel, info.Id)
                            confirmationModal := tview.NewModal().
                                SetText(confirmText).
                                AddButtons([]string{"Confirmer", "Annuler"}).
                                SetDoneFunc(func(confirmIndex int, confirmLabel string) {
                                    if confirmLabel == "Confirmer" {
                                        if buttonLabel == "Démarrer" {
                                            if err := startInstance(svc, info.Id); err != nil {
                                                fmt.Printf("Erreur lors du démarrage de %s: %v\n", info.Id, err)
                                            } else {
                                                fmt.Printf("Instance démarrée: %s\n", info.Id)
                                            }
                                        } else if buttonLabel == "Arrêter" {
                                            if err := stopInstance(svc, info.Id); err != nil {
                                                fmt.Printf("Erreur lors de l'arrêt de %s: %v\n", info.Id, err)
                                            } else {
                                                fmt.Printf("Instance arrêtée: %s\n", info.Id)
                                            }
                                        }
                                    }
                                    app.Stop()
                                })
                            app.SetRoot(confirmationModal, false).SetFocus(confirmationModal)
                        default:
                            app.Stop()
                        }
                    })
                app.SetRoot(actionModal, false).SetFocus(actionModal)
            }
        }(info))
    }

    list.SetBorder(true).SetTitle("Sélectionnez une instance").SetBorderPadding(1, 1, 1, 1)
    if err := app.SetRoot(list, true).SetFocus(list).Run(); err != nil {
        fmt.Printf("Erreur lors de l'exécution de l'application: %v\n", err)
    }
}