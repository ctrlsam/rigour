package types

import (
	"encoding/json"
	"time"

	"github.com/ctrlsam/rigour/pkg/crawler/fingerprint/plugins"
)

// Service represents a single service discovered on a host.
type Service struct {
	IP        string    `json:"ip" bson:"ip"`
	Port      int       `json:"port" bson:"port"`
	Protocol  string    `json:"protocol" bson:"protocol"`
	TLS       bool      `json:"tls" bson:"tls"`
	Transport string    `json:"transport" bson:"transport"`
	LastScan  time.Time `json:"last_scan" bson:"last_scan"`

	// Protocol-specific metadata fields - using plugin types
	HTTP          *plugins.ServiceHTTP          `json:"http,omitempty" bson:"http,omitempty"`
	HTTPS         *plugins.ServiceHTTPS         `json:"https,omitempty" bson:"https,omitempty"`
	SSH           *plugins.ServiceSSH           `json:"ssh,omitempty" bson:"ssh,omitempty"`
	RDP           *plugins.ServiceRDP           `json:"rdp,omitempty" bson:"rdp,omitempty"`
	SMB           *plugins.ServiceSMB           `json:"smb,omitempty" bson:"smb,omitempty"`
	RPC           *plugins.ServiceRPC           `json:"rpc,omitempty" bson:"rpc,omitempty"`
	MSSQL         *plugins.ServiceMSSQL         `json:"mssql,omitempty" bson:"mssql,omitempty"`
	Netbios       *plugins.ServiceNetbios       `json:"netbios,omitempty" bson:"netbios,omitempty"`
	Kafka         *plugins.ServiceKafka         `json:"kafka,omitempty" bson:"kafka,omitempty"`
	Oracle        *plugins.ServiceOracle        `json:"oracle,omitempty" bson:"oracle,omitempty"`
	MySQL         *plugins.ServiceMySQL         `json:"mysql,omitempty" bson:"mysql,omitempty"`
	SMTP          *plugins.ServiceSMTP          `json:"smtp,omitempty" bson:"smtp,omitempty"`
	SMTPS         *plugins.ServiceSMTPS         `json:"smtps,omitempty" bson:"smtps,omitempty"`
	LDAP          *plugins.ServiceLDAP          `json:"ldap,omitempty" bson:"ldap,omitempty"`
	Modbus        *plugins.ServiceModbus        `json:"modbus,omitempty" bson:"modbus,omitempty"`
	LDAPS         *plugins.ServiceLDAPS         `json:"ldaps,omitempty" bson:"ldaps,omitempty"`
	IMAP          *plugins.ServiceIMAP          `json:"imap,omitempty" bson:"imap,omitempty"`
	Rsync         *plugins.ServiceRsync         `json:"rsync,omitempty" bson:"rsync,omitempty"`
	Rtsp          *plugins.ServiceRtsp          `json:"rtsp,omitempty" bson:"rtsp,omitempty"`
	IMAPS         *plugins.ServiceIMAPS         `json:"imaps,omitempty" bson:"imaps,omitempty"`
	MQTT          *plugins.ServiceMQTT          `json:"mqtt,omitempty" bson:"mqtt,omitempty"`
	POP3          *plugins.ServicePOP3          `json:"pop3,omitempty" bson:"pop3,omitempty"`
	POP3S         *plugins.ServicePOP3S         `json:"pop3s,omitempty" bson:"pop3s,omitempty"`
	FTP           *plugins.ServiceFTP           `json:"ftp,omitempty" bson:"ftp,omitempty"`
	PostgreSQL    *plugins.ServicePostgreSQL    `json:"postgresql,omitempty" bson:"postgresql,omitempty"`
	VNC           *plugins.ServiceVNC           `json:"vnc,omitempty" bson:"vnc,omitempty"`
	Telnet        *plugins.ServiceTelnet        `json:"telnet,omitempty" bson:"telnet,omitempty"`
	Redis         *plugins.ServiceRedis         `json:"redis,omitempty" bson:"redis,omitempty"`
	SNMP          *plugins.ServiceSNMP          `json:"snmp,omitempty" bson:"snmp,omitempty"`
	NTP           *plugins.ServiceNTP           `json:"ntp,omitempty" bson:"ntp,omitempty"`
	IPSEC         *plugins.ServiceIPSEC         `json:"ipsec,omitempty" bson:"ipsec,omitempty"`
	Stun          *plugins.ServiceStun          `json:"stun,omitempty" bson:"stun,omitempty"`
	DNS           *plugins.ServiceDNS           `json:"dns,omitempty" bson:"dns,omitempty"`
	DHCP          *plugins.ServiceDHCP          `json:"dhcp,omitempty" bson:"dhcp,omitempty"`
	Echo          *plugins.ServiceEcho          `json:"echo,omitempty" bson:"echo,omitempty"`
	IPMI          *plugins.ServiceIPMI          `json:"ipmi,omitempty" bson:"ipmi,omitempty"`
	JDWP          *plugins.ServiceJDWP          `json:"jdwp,omitempty" bson:"jdwp,omitempty"`
	MinecraftJava *plugins.ServiceMinecraftJava `json:"minecraft-java,omitempty" bson:"minecraftjava,omitempty"`
}

// FromPluginService converts a plugins.Service to a types.Service, populating the appropriate
// protocol-specific metadata field based on the service protocol.
func FromPluginService(pluginSvc *plugins.Service, lastScan time.Time) *Service {
	svc := &Service{
		IP:        pluginSvc.IP,
		Port:      pluginSvc.Port,
		Protocol:  pluginSvc.Protocol,
		TLS:       pluginSvc.TLS,
		Transport: pluginSvc.Transport,
		LastScan:  lastScan,
	}

	// Unmarshal the raw JSON into the appropriate protocol-specific field
	switch pluginSvc.Protocol {
	case plugins.ProtoHTTP:
		var metadata plugins.ServiceHTTP
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.HTTP = &metadata
		}
	case plugins.ProtoHTTPS:
		var metadata plugins.ServiceHTTPS
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.HTTPS = &metadata
		}
	case plugins.ProtoSSH:
		var metadata plugins.ServiceSSH
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.SSH = &metadata
		}
	case plugins.ProtoRDP:
		var metadata plugins.ServiceRDP
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.RDP = &metadata
		}
	case plugins.ProtoSMB:
		var metadata plugins.ServiceSMB
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.SMB = &metadata
		}
	case plugins.ProtoRPC:
		var metadata plugins.ServiceRPC
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.RPC = &metadata
		}
	case plugins.ProtoMSSQL:
		var metadata plugins.ServiceMSSQL
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.MSSQL = &metadata
		}
	case plugins.ProtoNetbios:
		var metadata plugins.ServiceNetbios
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.Netbios = &metadata
		}
	case plugins.ProtoKafka:
		var metadata plugins.ServiceKafka
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.Kafka = &metadata
		}
	case plugins.ProtoOracle:
		var metadata plugins.ServiceOracle
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.Oracle = &metadata
		}
	case plugins.ProtoMySQL:
		var metadata plugins.ServiceMySQL
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.MySQL = &metadata
		}
	case plugins.ProtoSMTP:
		var metadata plugins.ServiceSMTP
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.SMTP = &metadata
		}
	case plugins.ProtoSMTPS:
		var metadata plugins.ServiceSMTPS
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.SMTPS = &metadata
		}
	case plugins.ProtoLDAP:
		var metadata plugins.ServiceLDAP
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.LDAP = &metadata
		}
	case plugins.ProtoModbus:
		var metadata plugins.ServiceModbus
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.Modbus = &metadata
		}
	case plugins.ProtoLDAPS:
		var metadata plugins.ServiceLDAPS
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.LDAPS = &metadata
		}
	case plugins.ProtoIMAP:
		var metadata plugins.ServiceIMAP
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.IMAP = &metadata
		}
	case plugins.ProtoRsync:
		var metadata plugins.ServiceRsync
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.Rsync = &metadata
		}
	case plugins.ProtoRtsp:
		var metadata plugins.ServiceRtsp
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.Rtsp = &metadata
		}
	case plugins.ProtoIMAPS:
		var metadata plugins.ServiceIMAPS
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.IMAPS = &metadata
		}
	case plugins.ProtoMQTT:
		var metadata plugins.ServiceMQTT
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.MQTT = &metadata
		}
	case plugins.ProtoPOP3:
		var metadata plugins.ServicePOP3
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.POP3 = &metadata
		}
	case plugins.ProtoPOP3S:
		var metadata plugins.ServicePOP3S
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.POP3S = &metadata
		}
	case plugins.ProtoFTP:
		var metadata plugins.ServiceFTP
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.FTP = &metadata
		}
	case plugins.ProtoPostgreSQL:
		var metadata plugins.ServicePostgreSQL
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.PostgreSQL = &metadata
		}
	case plugins.ProtoVNC:
		var metadata plugins.ServiceVNC
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.VNC = &metadata
		}
	case plugins.ProtoTelnet:
		var metadata plugins.ServiceTelnet
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.Telnet = &metadata
		}
	case plugins.ProtoRedis:
		var metadata plugins.ServiceRedis
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.Redis = &metadata
		}
	case plugins.ProtoSNMP:
		var metadata plugins.ServiceSNMP
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.SNMP = &metadata
		}
	case plugins.ProtoNTP:
		var metadata plugins.ServiceNTP
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.NTP = &metadata
		}
	case plugins.ProtoIPSEC:
		var metadata plugins.ServiceIPSEC
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.IPSEC = &metadata
		}
	case plugins.ProtoStun:
		var metadata plugins.ServiceStun
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.Stun = &metadata
		}
	case plugins.ProtoDNS:
		var metadata plugins.ServiceDNS
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.DNS = &metadata
		}
	case plugins.ProtoDHCP:
		var metadata plugins.ServiceDHCP
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.DHCP = &metadata
		}
	case plugins.ProtoEcho:
		var metadata plugins.ServiceEcho
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.Echo = &metadata
		}
	case plugins.ProtoIPMI:
		var metadata plugins.ServiceIPMI
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.IPMI = &metadata
		}
	case plugins.ProtoJDWP:
		var metadata plugins.ServiceJDWP
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.JDWP = &metadata
		}
	case plugins.ProtoMinecraftJava:
		var metadata plugins.ServiceMinecraftJava
		if err := json.Unmarshal(pluginSvc.Raw, &metadata); err == nil {
			svc.MinecraftJava = &metadata
		}
	}

	return svc
}
