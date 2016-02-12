package main

import (
  "fmt"
  "github.com/digitalocean/godo"
  "golang.org/x/oauth2"
  "log"
  "io/ioutil"
  "os/user"
  "strconv"
  "strings"
  "errors"
)


// FIXME: be able to configure the used SSH key
func main() {
  fmt.Println("running digital ocean setup script..")


  pat, err := ReadTokenFromConfigFile()
  if err != nil {
    log.Println(err)
    return
  }

  tokenSource := &TokenSource{
    AccessToken: pat,
  }
  oauthClient := oauth2.NewClient(oauth2.NoContext, tokenSource)
  client := godo.NewClient(oauthClient)

  droplets, err := DropletList(client)
  if err != nil {
    log.Println(err)
    return
  }

  log.Println(droplets)



  // err = RemoveAllDroplets(client)
  err = createMasterSlaveDroplets(client, 0)
  if err != nil {
    log.Println(err)
    return
  }

  droplets, err = DropletList(client)
  if err != nil {
    log.Println(err)
    return
  }

  log.Println(droplets)
}

func createMasterSlaveDroplets(client *godo.Client, slaveCount int) (err error) {

  sshKeys := []godo.DropletCreateSSHKey{godo.DropletCreateSSHKey{Fingerprint: "9e:6a:0b:3d:0a:d1:af:c6:7f:d3:00:aa:b3:a1:ed:dc"}}

  log.Println("creating master... ")
  _, err = createSmallDroplet(client, "master", sshKeys)
  if err != nil { return }

  for i := 0; i < slaveCount; i++ {
    slaveName := "slave" + strconv.Itoa(i)
    log.Println("creating slave with name " + slaveName)
    _, err = createSmallDroplet(client, slaveName, sshKeys)
    if err != nil { return }
  }
  return
}

func createSmallDroplet(client *godo.Client, dropletName string, sshKeys []godo.DropletCreateSSHKey) (*godo.Droplet, error) {

  // Docker 1.10.1 on 14.04 in San Francisco
  createRequest := &godo.DropletCreateRequest{
    Name:   dropletName,
    Region: "sfo1",
    Size:   "512mb",
    PrivateNetworking: true,
    SSHKeys: sshKeys,
    Image: godo.DropletCreateImage{
      Slug: "docker",
    },
  }

  newDroplet, _, err := client.Droplets.Create(createRequest)
  return newDroplet, err
}

func RemoveAllDroplets(client *godo.Client) error {
  droplets, err := DropletList(client)
  if err != nil {
    return err
  }

  for _, droplet := range droplets {
    _, err := client.Droplets.Delete(droplet.ID)

    if (err != nil) {
      return err
    }
  }
  return nil
}

type TokenSource struct {
AccessToken string
}

func DropletList(client *godo.Client) ([]godo.Droplet, error) {
  // create a list to hold our droplets
  list := []godo.Droplet{}

  // create options. initially, these will be blank
  opt := &godo.ListOptions{}
  for {
    droplets, resp, err := client.Droplets.List(opt)
    if err != nil {
      return nil, err
    }

    // append the current page's droplets to our list
    for _, d := range droplets {
      list = append(list, d)
    }

    // if we are at the last page, break out the for loop
    if resp.Links == nil || resp.Links.IsLastPage() {
      break
    }

    page, err := resp.Links.CurrentPage()
    if err != nil {
      return nil, err
    }

    // set the page we want for the next request
    opt.Page = page + 1
  }

  return list, nil
}

func ReadTokenFromConfigFile() (token string, err error) {
  usr, err := user.Current()
  if err != nil { return }

  bytes, err := ioutil.ReadFile(usr.HomeDir + "/.digitalOceanToken")
  if err != nil { return }

  return string(bytes), err
}

func (t *TokenSource) Token() (*oauth2.Token, error) {
  token := &oauth2.Token{
    AccessToken: t.AccessToken,
  }
  return token, nil
}


func GetTentacularDroplets(droplets []godo.Droplet) (master *godo.Droplet, slaves []godo.Droplet) {
  slaves = []godo.Droplet{}
  for _, droplet := range droplets {
    if IsMasterDroplet(&droplet) {
      master = &droplet
    } else if IsSlaveDroplet(&droplet) {
      slaves = append(slaves, droplet)
    }
  }

  return master, slaves
}

// FIXME: implement
func RunTentacularOnDroplets(master *godo.Droplet, slaves []godo.Droplet) (err error) {
  if master == nil {
    return errors.New("Missing master node.")
  }

  if len(slaves) == 0 {
    return errors.New("No slave nodes available.")
  }

  return nil
}

func IsMasterDroplet(droplet *godo.Droplet) bool {
  return strings.HasPrefix(droplet.Name, "master")
}

func IsSlaveDroplet(droplet *godo.Droplet) bool {
  return strings.HasPrefix(droplet.Name, "slave")
}

