package main

import (
	"time"
	_ "github.com/go-sql-driver/mysql"
	"fmt"
	"strconv"
	"os"
	"bufio"
	"encoding/csv"
	"strings"
	"database/sql"
	"bytes"
	"io"
)

func main() {
	var count1 int64
	var recordsCount int64
	file, err := os.Open("/home/siva/LatestAppOpenUsers_20170512_to_20171107.txt")
	defer file.Close()

	if err != nil {
		println(err)
	}

	dbConn := getDBConnection()
	dbConn.SetMaxOpenConns(10)
	defer dbConn.Close()
	err = dbConn.Ping()
	if err != nil {
		fmt.Println(err.Error())
	}
	// Start reading from the file with a reader.
	reader := bufio.NewReader(file)

	//outputfile1, err := os.Create("text1.txt")
	//if(err!=nil){
	//	fmt.Println("Not able to create a file")
	//}
	//defer outputfile1.Close()

	csvfile1, err := os.Create("resultplatform.csv")
	if(err!=nil){
		fmt.Println("Not able to create a csv file")
	}

	writer1 := csv.NewWriter(csvfile1)
	defer writer1.Flush()
	defer csvfile1.Close()

	limiter := time.Tick(time.Nanosecond * 333333333)

	var linecount int
	for {
		linecount =0
		var user platformUser
		lines := make([]string, 2000)
		query := ""
		for linecount<1000 {
			var buffer bytes.Buffer
			var line []byte
			line, _, err = reader.ReadLine()
			buffer.Write(line)
			println(buffer.String())
			// If we're just at the EOF, break
			if err != nil {
				if query=="" {
					fmt.Println("Final Number of records exported from the DB",recordsCount)
					os.Exit(1)
				} else {
					break
				}
			} else {
				uidString := string(line[:])
				uid :=uidString[0:16]
				lines = append(lines,uid)
				if linecount == 0 {
					query = query + "\"" + strings.TrimSpace(uid) + "\""
				} else {
					query = query + ",\"" + strings.TrimSpace(uid) + "\""
				}
				linecount++
			}
		}

		//fmt.Println("select * from platform_user where  hike_uid in ("+query+")")
		<-limiter
		rows1,err := dbConn.Query("select * from platform_user where  hike_uid in ("+query+")")
		if(err!=nil){
			fmt.Println(err)
		}

		for rows1.Next() {
			err := rows1.Scan(&user.ID,&user.HikeUID, &user.PlatformUID, &user.PlatformToken, &user.Msisdn,
				&user.HikeToken,&user.CreateTime,&user.UpdateTs, &user.Status)
			if(err!=nil) {
				fmt.Println(err.Error())
			}
			userCreateTime := strings.Split(user.CreateTime.String(),"+")
			userCrTime := userCreateTime[0]


			userUpdateTime := strings.Split(user.UpdateTs.String(),"+")
			userUpTime := userUpdateTime[0]


			msisdnReqd := user.Msisdn
			if strings.HasPrefix(msisdnReqd,"+9") {
				msisdnReqd=strings.Replace(msisdnReqd,"+9","1",1)
			} else if strings.HasPrefix(msisdnReqd,"+8") {
				msisdnReqd=strings.Replace(msisdnReqd,"+8","2",1)
			} else if strings.HasPrefix(msisdnReqd,"+7") {
				msisdnReqd=strings.Replace(msisdnReqd,"+7","3",1)
			} else {
				continue
			}

			count1++
			recordsCount++
			//outputfile1.WriteString(ToIntegerVal(count1)+"::"+user.HikeUID+"::"+user.PlatformUID+"::"+user.
			//	PlatformToken+"::+"+msisdnReqd+"::"+user.HikeToken+"::"+strings.TrimSpace(userCrTime)+"::"+strings.TrimSpace(userUpTime)+
			//	"::"+ToString(user.Status)+"\n")

			records1 := [][]string{
				{ToIntegerVal(count1),user.HikeUID,user.PlatformUID,user.PlatformToken,"+"+msisdnReqd,user.HikeToken,
					strings.TrimSpace(userCrTime),strings.TrimSpace(userUpTime), ToString(user.Status)},
			}

			for _, value := range records1 {
				err := writer1.Write(value)
				if(err!=nil){
					fmt.Println(err.Error())
					fmt.Println("Not able to write the records into csv file")
				}
			}
		}
		rows1.Close()
		fmt.Println("Number of records exported from the DB",recordsCount)
	}

	fmt.Println("Final Number of records exported from the DB",recordsCount)

	if err != io.EOF {
		fmt.Printf(" > Failed!: %v\n", err)
	}

}

func getDBConnection() *sql.DB{

	db, err := sql.Open("mysql", "platform:p1@tf0rmD1st@tcp(10.9.33.14:3306)/usersdb?parseTime=true")
	if(err!=nil){
		fmt.Println(err)
	}
	return db
}

func ToNullString(s string) sql.NullString {
	return sql.NullString{String : s, Valid : s != ""}
}

func ToIntegerVal(i int64) string {
	var valueInt string
	valueInt = strconv.FormatInt(int64(i), 10)
	return valueInt
}

func ToStringFromInt(i int) string {
	var valueInt string
	valueInt = strconv.Itoa(i)
	return valueInt
}

func ToString(s sql.NullString) string {
	var valInString string
	if(s.Valid) {
		valInString = s.String
		fmt.Println(valInString)
	} else {
		valInString = "NULL"
		fmt.Println(valInString)
	}
	return valInString
}

type platformUser struct {
	CreateTime    time.Time  `json:"create_time"`
	HikeToken     string `json:"hike_token"`
	HikeUID       string `json:"hike_uid"`
	ID            int64    `json:"id"`
	Msisdn        string `json:"msisdn"`
	PlatformToken string `json:"platform_token"`
	PlatformUID   string `json:"platform_uid"`
	Status        sql.NullString `json:"status"`
	UpdateTs      time.Time `json:"update_ts"`
}