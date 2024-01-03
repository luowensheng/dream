package demo

import (
	"fmt"
	"strconv"
	"github.com/luowensheng/dream"
)

var ALIEN_IMAGES = []string{
	"https://www.ox.ac.uk/sites/files/oxford/field/field_image_main/Aliens.jpg",
	"https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcRKK46hzdrS3iTHPxKQiGNq1F1p3Mhnp5yAs9_wbCc8GJCNjQ-Xp-cX5FxlJfVNn-s6Mzc&usqp=CAU",
	"https://scitechdaily.com/images/Alien-Holding-Planet-Earth-777x518.jpg?ezimgfmt=rs:350x233/rscb2/ngcb2/notWebP",
	"https://media-cldnry.s-nbcnews.com/image/upload/t_fit-760w,f_auto,q_auto:best/newscms/2017_06/1894401/170207-aliens-rhk-1646p.jpg",
	"https://media.npr.org/assets/img/2015/05/27/istock_000021797874small-ae052da70450a047e74266649594a03895311250-s1100-c50.jpg",
}

var BroadcastButtonClicked = "BroadcastButtonClicked"

func page() {

	numberOfImages := len(ALIEN_IMAGES)
	currentIndex := 0
	dream.LoadCSS("./assets/style.css")

	dream.El("button").Content("Broadcast").Broadcast(BroadcastButtonClicked, "click")

	dream.OnBroadcast(BroadcastButtonClicked, func() {
		fmt.Println("NEW BROADCAST!!!!!!!!!!!!!!!!")

	})

	dream.OnBroadcastWithParams(BroadcastButtonClicked, func(r dream.Record) {

		dream.ExecuteWithResponse("JSON.stringify(document.location)", func(output string) {
			fmt.Println("Current Document Location: ", output)
		})
	})

	imgRef := dream.El("img").
		Attr("class", "main-img").
		Attr("src", "/static/img.png")

	imgRef.On("click", func() {
		imgRef.ToggleClass("invisible")
	})

	dream.El("input").InnerRef(func(input *dream.ElementRef) {
		input.OnWithParams("keydown", func(params dream.Record) {
			// fmt.Println(params["key"])
			input.SetValue("Hacked!!")

		}, dream.Record{"key": "event.key"})
	})

	dream.El("div").
		Attr("style", "display:flex; flex-direction:row; height: 100%").
		Attr("class", "center").
		InnerRef(func(div *dream.ElementRef) {

			leftButtonRef := dream.El("button").Content("-").Attr("class", "click-button")

			h1Ref := dream.El("span").Content("0").Attr("class", "title")
			buttonClickedParams := dream.Record{"counter": h1Ref.GetTextContent()}

			rightButtonRef := dream.El("button").Content("+").Attr("class", "click-button")

			leftButtonRef.OnWithParams("click", func(params dream.Record) {
				counter, err := strconv.Atoi(params["counter"])
				if err != nil {
					fmt.Println(err)
					return
				}
				counter -= 1
				currentIndex = -1
				if currentIndex < 0 {
					currentIndex = numberOfImages - 1
				}
				imgRef.SetAttribute("src", ALIEN_IMAGES[currentIndex])
				h1Ref.SetTextContent(fmt.Sprintf("%d", counter))

			}, buttonClickedParams)

			rightButtonRef.OnWithParams("click", func(params dream.Record) {

				counter, err := strconv.Atoi(params["counter"])
				if err != nil {
					fmt.Println(err)
					return
				}
				counter += 1
				currentIndex += 1
				if currentIndex > numberOfImages-1 {
					currentIndex = 0
				}
				imgRef.SetAttribute("src", ALIEN_IMAGES[currentIndex])
				h1Ref.SetTextContent(fmt.Sprintf("%d", counter))

			}, buttonClickedParams)

		})

}

func Demo() {
	dream.CreateApp("Surla", 9090, page)
}
