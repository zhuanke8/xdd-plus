package models

//
//import (
//	"context"
//	"crypto/md5"
//	"crypto/rand"
//	"encoding/base64"
//	"encoding/hex"
//	"errors"
//	"fmt"
//	"github.com/StackExchange/wmi"
//	"net"
//	"sort"
//	"strings"
//)
//
//func GetPhysicalID() string {
//	var ids []string
//	if cpuinfo, err := getCPUInfo(); err != nil && len(cpuinfo) > 0 {
//		panic(err.Error())
//	} else {
//		ids = append(ids, cpuinfo[0].VendorID+cpuinfo[0].PhysicalID)
//	}
//	if mac, err := getMACAddress(); err != nil {
//		panic(err.Error())
//	} else {
//		ids = append(ids, mac)
//	}
//	sort.Strings(ids)
//	idsstr := strings.Join(ids, "|/|")
//	return GetMd5String(idsstr, true, true)
//}
//
//func getMACAddress() (string, error) {
//	netInterfaces, err := net.Interfaces()
//	if err != nil {
//		panic(err.Error())
//	}
//	mac, macerr := "", errors.New("无法获取到正确的MAC地址")
//	for i := 0; i < len(netInterfaces); i++ {
//		//fmt.Println(netInterfaces[i])
//		if (netInterfaces[i].Flags&net.FlagUp) != 0 && (netInterfaces[i].Flags&net.FlagLoopback) == 0 {
//			addrs, _ := netInterfaces[i].Addrs()
//			for _, address := range addrs {
//				ipnet, ok := address.(*net.IPNet)
//				//fmt.Println(ipnet.IP)
//				if ok && ipnet.IP.IsGlobalUnicast() {
//					// 如果IP是全局单拨地址，则返回MAC地址
//					mac = netInterfaces[i].HardwareAddr.String()
//					return mac, nil
//				}
//			}
//		}
//	}
//	return mac, macerr
//}
//
//type cpuInfo struct {
//	CPU        int32  `json:"cpu"`
//	VendorID   string `json:"vendorId"`
//	PhysicalID string `json:"physicalId"`
//}
//
//type win32_Processor struct {
//	Manufacturer string
//	ProcessorID  *string
//}
//
//func getCPUInfo() ([]cpuInfo, error) {
//	var ret []cpuInfo
//	var dst []win32_Processor
//	q := wmi.CreateQuery(&dst, "")
//	fmt.Println(q)
//	if err := wmiQuery(q, &dst); err != nil {
//		return ret, err
//	}
//
//	var procID string
//	for i, l := range dst {
//		procID = ""
//		if l.ProcessorID != nil {
//			procID = *l.ProcessorID
//		}
//
//		cpu := cpuInfo{
//			CPU:        int32(i),
//			VendorID:   l.Manufacturer,
//			PhysicalID: procID,
//		}
//		ret = append(ret, cpu)
//	}
//
//	return ret, nil
//}
//
//// WMIQueryWithContext - wraps wmi.Query with a timed-out context to avoid hanging
//func wmiQuery(query string, dst interface{}, connectServerArgs ...interface{}) error {
//	ctx := context.Background()
//	if _, ok := ctx.Deadline(); !ok {
//		ctxTimeout, cancel := context.WithTimeout(ctx, 3000000000) //超时时间3s
//		defer cancel()
//		ctx = ctxTimeout
//	}
//
//	errChan := make(chan error, 1)
//	go func() {
//		errChan <- wmi.Query(query, dst, connectServerArgs...)
//	}()
//
//	select {
//	case <-ctx.Done():
//		return ctx.Err()
//	case err := <-errChan:
//		return err
//	}
//}
//
////生成32位md5字串
//func GetMd5String(s string, upper bool, half bool) string {
//	h := md5.New()
//	h.Write([]byte(s))
//	result := hex.EncodeToString(h.Sum(nil))
//	if upper == true {
//		result = strings.ToUpper(result)
//	}
//	if half == true {
//		result = result[8:24]
//	}
//	return result
//}
//
////利用随机数生成Guid字串
//func UniqueId() string {
//	b := make([]byte, 48)
//	if _, err := rand.Read(b); err != nil {
//		return ""
//	}
//	return GetMd5String(base64.URLEncoding.EncodeToString(b), true, false)
//}
