package main

import (
    "os"
    log "github.com/Sirupsen/logrus"

    "github.com/armon/consul-api"
    flags "github.com/jessevdk/go-flags"
    
    "github.com/mitchellh/goamz/aws"
)

type Options struct {
    DataDir string `short:"d" long:"data-dir" required:"true" description:"data directory"`
    ZoneId  string `short:"z" long:"zone"     required:"true" description:"route53 zone id"`
    Name    string `short:"n" long:"name"     required:"true" description:"SRV record name"`
    TTL     int    `short:"t" long:"ttl"      required:"true" description:"TTL"`
    Debug   bool   `          long:"debug"                    description:"enable debug logging"`
    LogFile string `short:"l" long:"log-file"                 description:"JSON log file path"`
}

func main() {
    var opts Options
    
    _, err := flags.Parse(&opts)
    if err != nil {
        os.Exit(1)
    }
    
    if opts.Debug {
        // Only log the warning severity or above.
        log.SetLevel(log.DebugLevel)
    }
    
    if opts.LogFile != "" {
        // this fd will leak; we're not going to close it
        logFp, err := os.OpenFile(opts.LogFile, os.O_WRONLY | os.O_APPEND | os.O_CREATE, 0600)
        
        if err != nil {
            log.Fatalf("error opening %s: %v", opts.LogFile, err)
        }

        // log as JSON
        log.SetFormatter(&log.JSONFormatter{})
        
        // send output to file
        log.SetOutput(logFp)
    }
    
    consul, err := consulapi.NewClient(consulapi.DefaultConfig())
    
    if err != nil {
        log.Fatalf("unable to create consul client: %v", err)
    }
    
    awsAuth, err := aws.EnvAuth()

    if err != nil {
        log.Fatalf("unable to load AWS auth from environment: %v", err)
    }
    
    wrapper := NewLockWrapper(consul, opts.DataDir, "srv_recorder/" + opts.Name)
    updater := NewSrvUpdater(awsAuth, opts.ZoneId)
    
    if ! wrapper.loadSession() || ! wrapper.isSessionValid() {
        wrapper.createSession()
    }

    if wrapper.acquireLock() || wrapper.haveLock() {
        log.Debug("have lock")
        
        // retrieve the list of consul servers
        services, _, err := consul.Catalog().Service("consul", "", nil)
        
        if err != nil {
            log.Fatalf("unable to retrieve 'consul' service: %v", err)
        }
        
        srvRecord := SrvRecord{
            Name: opts.Name,
            TTL:  opts.TTL,
            Targets: make([]SrvTarget, len(services)),
        }
        
        for ind, value := range services {
            // warning opinionated: Node name should be resolvable.  According
            // to the spec it should also be an A or AAAA record…
            
            // value.ServicePort returns the server's RPC address (8300 by
            // default), but joining the cluster requires using the server's
            // serf_lan port (8301 by default).  The serf_lan port isn't exposed
            // via the catalog, however.  Guess this'll just be more
            // opinionated, for now.
            srvRecord.Targets[ind] = SrvTarget{
                Priority: 10,
                Weight:   10,
                Port:     8301, // see above
                Target:   value.Address,
            }
        }
        
        err = updater.UpdateRecord(&srvRecord)
        
        if err != nil {
            log.Fatalf("unable to update record: %v", err)
        }
    } else {
        log.Info("unable to acquire lock")
    }
}
