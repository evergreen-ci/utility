package utility

import (
	"io/ioutil"
	"net"

	"github.com/pkg/errors"
)

func GetPublicIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", errors.Wrap(err, "establishing outbound connection")
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	_, cidr, _ := net.ParseCIDR("172.16.0.0/12")
	if cidr.Contains(localAddr.IP.To4()) {
		client := GetHTTPClient()
		defer PutHTTPClient(client)

		resp, err := client.Get("http://169.254.169.254/latest/meta-data/public-ipv4")
		if err != nil {
			return localAddr.IP.To4().String(), errors.Wrap(err, "accessing metadata service")
		}
		defer resp.Body.Close()

		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return localAddr.IP.To4().String(), errors.Wrap(err, "reading response body")
		}

		return string(data), nil
	}

	return localAddr.IP.To4().String(), nil
}
