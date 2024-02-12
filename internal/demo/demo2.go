package demo

import (
	"fmt"

	"github.com/luowensheng/dream"
)

func App2(page *dream.PageContext) {

	// page.El("button").Content("Item").Broadcast("btn-click", "click")

	// ul := page.El("ul").Content("Nothing")

	// page.OnBroadcast("btn-clicked", func() {
	// 	ul.SetInnerHTML("")
	// })


	button := page.El("button").Content("Click")
	ref := page.El("div")
	ref.Inner(func() {
		for i := 0; i < 10; i++ {
			page.El("div").Content(fmt.Sprintf("element-%d", i))
		}
	})

	button.On("click", func() {
		ref.UpdateInner(func(newContext *dream.PageContext) {
			// newContext.El("div").
			// 	Inner(func() {
					for i := 0; i < 100; i++ {
						el := newContext.El("div").
							Content(fmt.Sprintf("new-element-%d", i))

						el.On("click", func() {
								fmt.Println("hello")
								el.SetTextContent("AHAHAHAH")
							})
					}
				// })
		})
	})

	// fmt.Println(manager.FirstElement.String())

	// os.Exit(0)

}
