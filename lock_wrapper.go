package main

import (
    "os"
    "fmt"
    log "github.com/Sirupsen/logrus"
    "path/filepath"
    "encoding/json"
    "github.com/armon/consul-api"
)

type LockWrapper struct {
    consul  *consulapi.Client
    sessionPath string
    
    session     consulapi.SessionEntry
    keyPath     string
}

func NewLockWrapper(consul *consulapi.Client, dataDir string) (*LockWrapper) {
    agentInfo, err := consul.Agent().Self()
    
    if err != nil {
        log.Fatal("can't get agent info: ", err)
    }

    updater := &LockWrapper{
        consul:      consul,
        sessionPath: filepath.Join(dataDir, "session.json"),
        keyPath:     fmt.Sprintf("consul.io/srv_recorder/%s/leader", agentInfo["Config"]["Datacenter"]),
    }
    
    return updater
}

// creates a new session
func (self *LockWrapper) createSession() {
    log.Info("creating session")
    
    sessionId, _, err := self.consul.Session().Create(
        &consulapi.SessionEntry{
            Name: "consul-srv-updater",
        },
        nil,
    )
    
    if err != nil {
        log.Fatalf("unable to create session: %v", err)
    }
    
    session, _, err := self.consul.Session().Info(sessionId, nil)

    if err != nil {
        log.Fatalf("unable to retrieve session: %v", err)
    }
    
    self.session = *session
    
    err = self.storeSession()
    
    if err != nil {
        log.Fatal("unable to store session")
    }
}

// destroys an existing session, and also removes the session file
func (self *LockWrapper) destroySession() {
    log.Info("destroying session")

    self.consul.Session().Destroy(self.session.ID, nil)
    
    err := os.Remove(self.sessionPath)
    
    if err != nil {
        // not fatal.  just unfortunate
        log.Warnf("unable to remove %s: %v", self.sessionPath, err)
    }
}

// stores session to file, returning an error if unable
func (self *LockWrapper) storeSession() error {
    log.Debug("storing session")
    
    ofp, err := os.OpenFile(self.sessionPath, os.O_WRONLY | os.O_TRUNC | os.O_CREATE, 0600)
    
    if err != nil {
        return err
    }
    
    defer ofp.Close()
    
    encoder := json.NewEncoder(ofp)
    
    return encoder.Encode(&self.session)
}

// loads the session from the file, if it exists
func (self *LockWrapper) loadSession() bool {
    if _, err := os.Stat(self.sessionPath); err == nil {
        log.Debugf("loading session from %s", self.sessionPath)
        
        // file exists; deserialize
        ifp, err := os.Open(self.sessionPath)
        
        if err != nil {
            log.Fatalf("unable to open %s: %v", self.sessionPath, err)
        }
        
        // ensure we close the file
        defer ifp.Close()
        
        decoder := json.NewDecoder(ifp)
        
        err = decoder.Decode(&self.session)
        
        if err != nil {
            log.Warnf("unable to decode %s: %v", self.sessionPath, err)
        } else {
            log.WithFields(log.Fields{
                "session": self.session.ID,
            }).Debug("loaded session")
        }
    } else {
        log.Warnf("no session file at %s", self.sessionPath)
    }
    
    return self.session.ID != ""
}

// tests that the session is valid
func (self *LockWrapper) isSessionValid() bool {
    log.WithFields(log.Fields{
        "session": self.session.ID,
    }).Debug("validating session")
    
    // validate session; returns nil if session is invalid
    session, _, err := self.consul.Session().Info(self.session.ID, nil)

    if err != nil {
        log.Fatalf("unable to retrieve session: %v", err)
    }
    
    isValid := session != nil
    
    log.WithFields(log.Fields{
        "session": self.session.ID,
    }).Debugf("session is valid? %t", isValid)
    
    return isValid
}

func (self *LockWrapper) haveLock() bool {
    log.WithFields(log.Fields{
        "key": self.keyPath,
        "session": self.session.ID,
    }).Debug("checking for lock")
    
    kvp, _, err := self.consul.KV().Get(self.keyPath, nil)
    
    if err != nil {
        log.Fatalf("unable to get key %s: %v", self.keyPath, err)
    }

    return kvp.Session == self.session.ID
}

func (self *LockWrapper) acquireLock() bool {
    log.WithFields(log.Fields{
        "key": self.keyPath,
        "session": self.session.ID,
    }).Debug("acquiring lock")

    kvp := &consulapi.KVPair{
        Key: self.keyPath,
        Session: self.session.ID,
    }

    acquired, _, err := self.consul.KV().Acquire(kvp, nil)
    
    if err != nil {
        log.Fatalf("error acquiring lock: %v", err)
    }

    log.WithFields(log.Fields{
        "key": self.keyPath,
        "session": self.session.ID,
    }).Debugf("acquired lock? %t", acquired)

    return acquired
}
