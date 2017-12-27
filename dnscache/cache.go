package dnscache

import (
	"time"

	"github.com/go-redis/redis"
)

var instance *redis.Client = nil

func GetInstance() *redis.Client {
	if instance == nil {
		host := "127.0.0.1"
		port := "6379"
		addr := host + ":" + port
		instance = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "",
			DB:       0, // use default DB
		})

		_, err := instance.Ping().Result()
		if err != nil {
			panic(err)
		}
	}

	return instance
}

func IPv4Preffix() string {
	return "ipv4:"
}
func IPv6Preffix() string {
	return "ipv6:"
}

func AddDomainIPv4(domain, ip string, expiration int) error {
	err := GetInstance().Set(IPv4Preffix()+domain, ip, 0).Err()
	if err != nil {
		return err
	}

	// set expiration time (in seconds)
	if expiration > 0 {
		expDuration := time.Duration(expiration) * time.Second
		err = GetInstance().Expire(IPv4Preffix()+domain, expDuration).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func AddDomainIPv6(domain, ip string, expiration int) error {
	err := GetInstance().Set(IPv6Preffix()+domain, ip, 0).Err()
	if err != nil {
		return err
	}

	// set expiration time (in seconds)
	if expiration > 0 {
		expDuration := time.Duration(expiration) * time.Second
		err = GetInstance().Expire(IPv6Preffix()+domain, expDuration).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

func DeleteDomainIPv4(domain string) error {
	err := GetInstance().Del(IPv4Preffix() + domain).Err()
	if err != nil {
		return err
	}

	return nil
}

func DeleteDomainIPv6(domain string) error {
	err := GetInstance().Del(IPv6Preffix() + domain).Err()
	if err != nil {
		return err
	}

	return nil
}

func GetDomainIPv4(domain string) (string, error) {
	ip, err := GetInstance().Get(IPv4Preffix() + domain).Result()
	return ip, err
}

func GetDomainIPv6(domain string) (string, error) {
	ip, err := GetInstance().Get(IPv6Preffix() + domain).Result()
	return ip, err
}

func GetTTLDomainIPv4(domain string) (int64, error) {
	t, err := GetInstance().TTL(IPv4Preffix() + domain).Result()
	return int64(t), err
}

func GetTTLDomainIPv6(domain string) (int64, error) {
	t, err := GetInstance().TTL(IPv6Preffix() + domain).Result()
	return int64(t), err
}

func Flush() error {
	return GetInstance().FlushDB().Err()
}
