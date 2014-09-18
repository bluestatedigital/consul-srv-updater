package main

import (
    // "github.com/mitchellh/goamz/route53"
    
    "os"
    "fmt"
    "log"
    "path/filepath"
    "encoding/json"
    "github.com/armon/consul-api"
)

type SrvUpdater struct {
    consul  *consulapi.Client
    sessionPath string
    
    session     consulapi.SessionEntry
    keyPath     string
}

func NewSrvUpdater(consul *consulapi.Client, dataDir string) (*SrvUpdater) {
    agentInfo, err := consul.Agent().Self()
    
    if err != nil {
        log.Fatal("can't get agent info ", err)
    }

    updater := &SrvUpdater{
        consul:      consul,
        sessionPath: filepath.Join(dataDir, "session.json"),
        keyPath:     fmt.Sprintf("consul.io/srv_recorder/%s/leader", agentInfo["Config"]["Datacenter"]),
    }
    
    return updater
}

// creates a new session
func (self *SrvUpdater) createSession() {
    sessionId, _, err := self.consul.Session().Create(
        &consulapi.SessionEntry{
            Name: "consul-srv-updater",
        },
        nil,
    )
    
    if err != nil {
        log.Fatal("unable to create session: ", err)
    }
    
    session, _, err := self.consul.Session().Info(sessionId, nil)

    if err != nil {
        log.Fatal("unable to retrieve session: ", err)
    }
    
    self.session = *session
    
    err = self.storeSession()
    
    if err != nil {
        log.Fatal("unable to store session")
    }
}

// destroys an existing session, and also removes the session file
func (self *SrvUpdater) destroySession() {
    self.consul.Session().Destroy(self.session.ID, nil)
    
    err := os.Remove(self.sessionPath)
    
    if err != nil {
        // not fatal.  just unfortunate
        log.Print("unable to remove ", self.sessionPath, err)
    }
}

// stores session to file, returning an error if unable
func (self *SrvUpdater) storeSession() error {
    ofp, err := os.OpenFile(self.sessionPath, os.O_WRONLY | os.O_TRUNC | os.O_CREATE, 0600)
    
    if err != nil {
        return err
    }
    
    defer ofp.Close()
    
    encoder := json.NewEncoder(ofp)
    
    return encoder.Encode(&self.session)
}

// loads the session from the file, if it exists
func (self *SrvUpdater) loadSession() bool {
    if _, err := os.Stat(self.sessionPath); err == nil {
        // file exists; deserialize
        ifp, err := os.Open(self.sessionPath)
        
        if err != nil {
            log.Fatal("unable to open", self.sessionPath, err)
        }
        
        // ensure we close the file
        defer ifp.Close()
        
        decoder := json.NewDecoder(ifp)
        
        err = decoder.Decode(&self.session)
        
        if err != nil {
            log.Print("unable to decode", err)
        }
    }
    
    return self.session.ID != ""
}

// tests that the session is valid
func (self *SrvUpdater) isSessionValid() bool {
    // validate session; returns nil if session is invalid
    session, _, err := self.consul.Session().Info(self.session.ID, nil)

    if err != nil {
        log.Fatal("unable to retrieve session: ", err)
    }
    
    return session != nil
}

func (self *SrvUpdater) haveLock() bool {
    kvp, _, err := self.consul.KV().Get(self.keyPath, nil)
    
    if err != nil {
        log.Fatal("unable to get key ", self.keyPath, err)
    }

    return kvp.Session == self.session.ID
}

func (self *SrvUpdater) acquireLock() bool {
    kvp := &consulapi.KVPair{
        Key: self.keyPath,
        Session: self.session.ID,
    }

    acquired, _, err := self.consul.KV().Acquire(kvp, nil)
    
    if err != nil {
        log.Fatal("unable to acquire lock ", err)
    }
    
    return acquired
}

func (self *SrvUpdater) Update() error {
    if ! self.loadSession() || ! self.isSessionValid() {
        self.createSession()
    }

    if self.acquireLock() || self.haveLock() {
        log.Print("can do some stuff")
    } else {
        log.Print("unable to lock key")
    }
    
    return nil
}
