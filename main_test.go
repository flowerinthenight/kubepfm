package main

import (
	"strings"
	"testing"
)

type check struct {
	target  string
	ctx     string
	rctype  string
	name    string
	ns      string
	ports   string
	address string
}

func Test_parse(t *testing.T) {
	var args []string
	args, _, _, _ = parse("8080:1222")
	if args != nil {
		t.Fatal("should be nil")
	}

	for _, v := range []check{
		{"name:8080:1222", "", "pod", "name", "default", "8080:1222", ""},
		{"deployment/name:8080:1222", "", "deployment", "deployment/name", "default", "8080:1222", ""},
		{"system:service/name:8080:1222", "", "service", "service/name", "system", "8080:1222", ""},
		{"ns=system:service/name:8080:1222", "", "service", "service/name", "system", "8080:1222", ""},
		{"ctx=minikube:ns=system:deployment/name:8080:1222", "minikube", "deployment", "deployment/name", "system", "8080:1222", ""},
		{"ctx=minikube:ns=system:deployment/name:127.0.0.2:8080:1222", "minikube", "deployment", "deployment/name", "system", "8080:1222", "127.0.0.2"},
		{"ctx=minikube:system:name:8080:1222", "minikube", "pod", "name", "system", "8080:1222", ""},
	} {
		args, ctx, name, ports := parse(v.target)
		if ctx != v.ctx {
			t.Fatalf("[%v] context should be %v, got %v", v.target, v.ctx, ctx)
		}
		if args[1] != v.rctype {
			t.Fatalf("[%v] rctype should be %v, got %v", v.target, v.rctype, args[1])
		}
		if name != v.name {
			t.Fatalf("[%v] name should be %v, got %v", v.target, v.name, name)
		}
		for _, a := range args {
			if strings.HasPrefix(a, "--namespace=") {
				vv := strings.Split(a, "=")
				if vv[1] != v.ns {
					t.Fatalf("[%v] ns should be %v, got %v", v.target, v.ns, vv[1])
				}
			}
		}
		if ports != v.ports {
			t.Fatalf("[%v] ports should be %v, got %v", v.target, v.ports, ports)
		}
	}
}
