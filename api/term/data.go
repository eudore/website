package term

import (
	"fmt"
	"time"
)

/*
PostgreSQL Begin
CREATE SEQUENCE seq_term_user_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_term_user(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_term_user_id'),
	"name" VARCHAR(32) NOT NULL,
	"password" VARCHAR(64) DEFAULT '',
	"salf" VARCHAR(10) DEFAULT '',
	"publickey" VARCHAR(512)  DEFAULT ''
);

CREATE SEQUENCE seq_term_host_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_term_host(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_term_host_id'),
	"name" VARCHAR(32) NOT NULL,
	"protocol" VARCHAR(32) NOT NULL,
	"addr" VARCHAR(32) NOT NULL,
	"user" VARCHAR(32) NOT NULL,
	"password" VARCHAR(32) DEFAULT '',
	"privatekey" VARCHAR(2048) DEFAULT ''
);

CREATE TABLE tb_term_user_host(
	"userid" INTEGER,
	"hostid" INTEGER,
	"time" TIMESTAMP DEFAULT (now()),
	PRIMARY KEY("userid", "hostid")
);
CREATE VIEW vi_term_user_host AS
	SELECT u.id  "userid",u.name "username",u.password "userpassword",u.salf "usersalf",u.publickey "userpublickey",h.id "hostid",h.name "hostname",h.protocol,h.addr,h.user "hostuser",h.password as "hostpassrod",h.privatekey "hostprivatekey",uh.TIME "granttime"
	FROM tb_term_user AS u JOIN tb_term_user_host AS uh ON uh.userid = u.id JOIN tb_term_host AS h ON uh.hostid = h.id;

CREATE SEQUENCE seq_term_hostgroup_id INCREMENT by 1 MINVALUE 1 START 1;
CREATE TABLE tb_term_hostgroup(
	"id" INTEGER PRIMARY KEY DEFAULT nextval('seq_term_hostgroup_id'),
	"name" VARCHAR(32) NOT NULL,
	"description" VARCHAR(256) DEFAULT ''
);

CREATE TABLE tb_term_user_hostgroup(
	"userid" INTEGER,
	"groupid" INTEGER,
	"time" TIMESTAMP DEFAULT (now()),
	PRIMARY KEY("userid", "groupid")
);
CREATE VIEW vi_term_user_hostgroup AS
	SELECT u.id  "userid",u.name "username",u.password,u.salf,u.publickey,g.id "groupid",g.name "groupname",g.description "groupdescription",uh.TIME "granttime"
	FROM tb_term_user AS u JOIN tb_term_user_hostgroup AS uh ON uh.userid = u.id JOIN tb_term_hostgroup AS g ON uh.groupid = g.id;

CREATE TABLE tb_term_hostgroup_host(
	"groupid" INTEGER,
	"hostid" INTEGER,
	"time" TIMESTAMP DEFAULT (now()),
	PRIMARY KEY("hostid", "groupid")
);
CREATE VIEW vi_term_host_hosgroupt AS
	SELECT g.id "groupid",g.name "groupname",g.description,h.id "hostid",h.name "hostname",h.protocol,h.addr,h.user,h.password,h.privatekey,hg.TIME "granttime"
	FROM tb_term_host AS h JOIN tb_term_hostgroup_host AS hg ON hg.hostid = h.id JOIN tb_term_hostgroup AS g ON hg.groupid = g.id;

PostgreSQL End
*/

type Host struct {
	ID         int    `alias:"id" json:"id"`
	Name       string `alias:"name" json:"name"`
	Protocol   string `alias:"protocol" json:"protocol"`
	Addr       string `alias:"addr" json:"addr"`
	User       string `alias:"user" json:"user"`
	Password   string `alias:"password" json:"password" masking:"0"`
	PrivateKey string `alias:"privatekey" json:"privatekey" masking:"0"`
}
type User struct {
	ID        int    `alias:"id" json:"id"`
	Name      string `alias:"name" json:"name"`
	Password  string `alias:"password" json:"password" masking:"0"`
	Salf      string `alias:"salf" json:"salf"`
	PublicKey string `alias:"publickey" json:"publickey" masking:"0"`
}
type ViewUserHost struct {
	Userid         int       `alias:"userid" json:"userid"`
	Username       string    `alias:"username" json:"username"`
	Userpassword   string    `alias:"userpassword" json:"userpassword" masking:"0"`
	Usersalf       string    `alias:"usersalf" json:"usersalf"`
	Userpublickey  string    `alias:"userpublickey" json:"userpublickey" masking:"0"`
	Hostid         int       `alias:"hostid" json:"hostid"`
	Hostname       string    `alias:"hostname" json:"hostname"`
	Protocol       string    `alias:"protocol" json:"protocol"`
	Addr           string    `alias:"addr" json:"addr"`
	Hostuser       string    `alias:"hostuser" json:"hostuser"`
	Hostpassrod    string    `alias:"hostpassrod" json:"hostpassrod"`
	Hostprivatekey string    `alias:"hostprivatekey" json:"hostprivatekey"`
	Granttime      time.Time `alias:"granttime" json:"granttime"`
}

type ViewUserHostgroup struct {
	Userid           int       `alias:"userid" json:"userid"`
	Username         string    `alias:"username" json:"username"`
	Password         string    `alias:"password" json:"password" masking:"0"`
	Salf             string    `alias:"salf" json:"salf"`
	Publickey        string    `alias:"publickey" json:"publickey" masking:"0"`
	Groupid          int       `alias:"groupid" json:"groupid"`
	Groupname        string    `alias:"groupname" json:"groupname"`
	Groupdescription string    `alias:"groupdescription" json:"groupdescription"`
	Granttime        time.Time `alias:"granttime" json:"granttime"`
}

type ViewHostHostgroup struct {
	Groupid     int       `alias:"groupid" json:"groupid"`
	Groupname   string    `alias:"groupname" json:"groupname"`
	Description string    `alias:"description" json:"description"`
	Hostid      int       `alias:"hostid" json:"hostid"`
	Hostname    string    `alias:"hostname" json:"hostname"`
	Protocol    string    `alias:"protocol" json:"protocol"`
	Addr        string    `alias:"addr" json:"addr"`
	User        string    `alias:"user" json:"user"`
	Password    string    `alias:"password" json:"password" masking:"0"`
	Privatekey  string    `alias:"privatekey" json:"privatekey" masking:"0"`
	Granttime   time.Time `alias:"granttime" json:"granttime"`
}

func (host Host) Format(format string) string {
	return fmt.Sprintf(format, host.Protocol, host.Addr, host.User)
}
