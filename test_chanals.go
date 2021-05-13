package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"time"
)

//this func generates numbers from 1 to 10 at intervals of 0.5 to 3 seconds and transmits them to the channel
func randomGenerator(channelRandomNumber chan<- float64, channelTimeMS chan<- int64, closeChannel <-chan bool) {
	var result float64

	generationPeriodMsec := math.Round((rand.Float64()/1*(3.0-0.5) + 0.5) * 1000)
	generationPeriodSec := time.Duration(int64(generationPeriodMsec) * int64(time.Millisecond))
	timeGenerator := time.NewTicker(generationPeriodSec)
	now := time.Now()
	nanos := now.UnixNano()
	millis := nanos / 1000000
	for {
		select {
		case <-closeChannel:
			close(channelRandomNumber)
			close(channelTimeMS)
			timeGenerator.Stop()
			break
		case <-timeGenerator.C:
			rand.Seed(time.Now().UnixNano())
			generationPeriodMsec = math.Round((rand.Float64()/1*(3.0-0.5) + 0.5) * 1000)
			generationPeriodSec = time.Duration(int64(generationPeriodMsec) * int64(time.Millisecond))
			result = rand.Float64()/1*(10.0-1) + 1
			now = time.Now()
			nanos = now.UnixNano()
			millis = nanos / 1000000
			fmt.Println(millis)
			fmt.Println(generationPeriodSec)
			channelRandomNumber <- result
			channelTimeMS <- millis
			timeGenerator.Reset(generationPeriodSec)
		}
	}
}

func printResult(actualCountNumbers int64, sumNumbers float64, average float64, movingAverage float64, slice1 []float64, data map[int64]float64) {
	fmt.Println("Количество сгенерированных элементов\t", actualCountNumbers)
	fmt.Print("Сумма всех сгенерированных элементов\t")
	fmt.Printf("%.2f\n", sumNumbers)
	fmt.Print("Среднее значение\t")
	fmt.Printf("%.2f\n", average)
	fmt.Print("Скользящее среднее\t")
	fmt.Printf("%.2f\n", movingAverage)
	fmt.Println(slice1)
	fmt.Println(data)
}

func main() {

	channelRandomGenerator := make(chan float64)
	channelTimeMS := make(chan int64)
	channelEndProgram := make(chan os.Signal, 1)
	channelCloseRandomGenerator := make(chan bool)

	var sumNumbers, average, movingAverage float64
	var actualCountNumbers int64 = 0
	var periodMovingAverage = 20000
	var periodPrintControl = time.Duration(5)
	var supportElement = periodMovingAverage / 2000
	timeoutPrintControl := time.NewTicker(periodPrintControl * time.Second)
	var timeMS int64
	structDate := make(map[int64]float64)
	slice1 := make([]float64, supportElement)

	signal.Notify(channelEndProgram, os.Interrupt)

	go randomGenerator(channelRandomGenerator, channelTimeMS, channelCloseRandomGenerator)

	//this is the main goroutine. this goroutine waits for a number to be generated in the channel, prints out the calculated values and

	for {
		select {
		case randomNumber, ok := <-channelRandomGenerator:
			if !ok {
				fmt.Println("Канал генерации и передачи данных закрыт")
				printResult(actualCountNumbers, sumNumbers, average, movingAverage, slice1, structDate)
				os.Exit(123)

			} else {
				timeMS = <-channelTimeMS
				structDate[timeMS] = randomNumber
				actualCountNumbers += 1
				slice1[len(slice1)-1] = randomNumber
				sumNumbers += randomNumber
				average = sumNumbers / float64(actualCountNumbers)
				sum := 0.0
				for _, ii := range slice1 {
					sum += ii
				}
				movingAverage = sum / float64(supportElement)
				slice1 = slice1[1:(len(slice1))]
				slice1 = append(slice1, 0)
			}
		case <-timeoutPrintControl.C:
			fmt.Println("==============\tКонтроль\t==============")
			printResult(actualCountNumbers, sumNumbers, average, movingAverage, slice1, structDate)
			fmt.Println()
			fmt.Println()
		case <-channelEndProgram:
			fmt.Println("Пользователь завершил выполнение программы")
			timeoutPrintControl.Stop()
			channelCloseRandomGenerator <- true
		}
	}
}
