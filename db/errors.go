package db

import (
	"errors"
	"github.com/sirupsen/logrus"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func UppendErrorWithPath(err error) {
	if strings.Contains(err.Error(), "ERROR: cached plan must not change result") {
		os.Exit(1)
	}
	badway := ""
	for stepCount := 1; stepCount <= 10; stepCount++ {
		// получаем через рантайм указатель на положение stepCount в цепочке шаг наверх, а так же строку вызова внутри функции.
		pc, _, line, _ := runtime.Caller(stepCount)
		// получаем через указатель полный путь к функции записанный на этом шагу
		fullFuncPath := runtime.FuncForPC(pc).Name()
		// дробим строку через точку, что бы получить массив строк, в котором последним элементом останется имя функции.
		splitedFuncPath := strings.Split(fullFuncPath, ".")
		// выбираем имя функции из массива в новую переменную, для лёгкого чтения дальнейших строк.
		funcName := splitedFuncPath[len(splitedFuncPath)-1]
		// если за stepCount шагов мы дошли до роута, то заканчиваем сбор имён в цепочке функций.
		if funcName == "ServeHTTP" {
			break
		} else {
			// исключаем пустые имена функций, что периодически могут появляться если шаги ушли достаточно далеко.
			if funcName != "" {
				// сохраняем/дописываем в crudLog.FuncPath имя и строку вызова функции.
				badway += funcName + "(" + strconv.Itoa(line) + ") | "
			}
		}
	}
	if err != nil {
		uppendError(badway, err.Error())
	} else {
		uppendError(badway, errors.New("передали пустую ошибку").Error())
	}
}

func uppendError(place string, newerr string) {
	var errLog ErrLogs
	logrus.Error("func uppendError( ", place, newerr)

	if err := Database.Db.Where("place = ? and error = ?", place, newerr).First(&errLog).Error; err != nil {
		if err != nil {
			logrus.Warn(err)
		}
		errLog.Place = place
		errLog.Error = newerr
		errLog.Count = 1
		Database.Db.Create(&errLog)
	} else {
		errLog.Count = errLog.Count + 1
		err = Database.Db.Model(&errLog).Update("count", errLog.Count).Error
		if err != nil {
			logrus.Warn(err)
		}

	}
}
