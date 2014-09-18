package main

import (
    "os"
    "github.com/armon/consul-api"
    flags "github.com/jessevdk/go-flags"
)

type Options struct {
    DataDir string `short:"d" long:"data-dir" description:"data directory"`
}

func main() {
    var opts Options
    
    _, err := flags.Parse(&opts)
    if err != nil {
        os.Exit(1)
    }

    client, err := consulapi.NewClient(consulapi.DefaultConfig())
    
    if err != nil {
        panic(err)
    }
    
    updater := NewSrvUpdater(client, opts.DataDir)
    
    updater.Update()
}
