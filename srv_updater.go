package main

import (
    log "github.com/Sirupsen/logrus"
    "fmt"
    
    Route53 "github.com/mitchellh/goamz/route53"
    AWS "github.com/mitchellh/goamz/aws"
)

type SrvTarget struct {
    Priority int
    Weight   int
    Port     int
    Target   string // grr, per the spec this needs to be an A or AAAA record
}

// _service._proto.name. TTL class SRV priority weight port target.
type SrvRecord struct {
    Name    string
    TTL     int
    Targets []SrvTarget
}

type SrvUpdater struct {
    zoneId  string
    route53 *Route53.Route53
}

func NewSrvUpdater(auth AWS.Auth, zoneId string) *SrvUpdater {
    srvUpdater := &SrvUpdater{
        zoneId:  zoneId,
        route53: Route53.New(auth, AWS.USEast),
    }
    
    return srvUpdater
}

func (self *SrvUpdater) UpdateRecord(rec *SrvRecord) error {
    log.Infof("updating %s SRV record with %d targets", rec.Name, len(rec.Targets))
    
    crrsReq := &Route53.ChangeResourceRecordSetsRequest{
        Changes: make([]Route53.Change, 1),
    }
    
    crrsReq.Changes[0] = Route53.Change{
        Action: "UPSERT",
        Record: Route53.ResourceRecordSet{
            Name: rec.Name,
            Type: "SRV",
            TTL: rec.TTL,
            Records: make([]string, len(rec.Targets)),
        },
    }
    
    for ind, target := range rec.Targets {
        crrsReq.Changes[0].Record.Records[ind] = fmt.Sprintf(
            "%d %d %d %s",
            target.Priority,
            target.Weight,
            target.Port,
            target.Target,
        )
    }
    
    resp, err := self.route53.ChangeResourceRecordSets(self.zoneId, crrsReq)
    
    if err != nil {
        log.Debugf("unable to update record: %v", err)
    } else {
        log.Debugf("change status: %s", resp.ChangeInfo.Status)
    }
    
    return err
}
