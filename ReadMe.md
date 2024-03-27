# A Rules based Pi Traffic Light written in go
#
I picked this project to learn both go and a rules engine.   It uses a "off the shelf" stoplight signal that simply plugs into the pi.   The "push to walk" button came in a kit with buttons and is also available separately.

Pi Traffic Light on Amazon
https://www.amazon.com/dp/B00RIIGD30?psc=1&ref=ppx_pop_dt_b_product_details

And using this go library
https://github.com/stianeikeland/go-rpio

# Go Code
The code is pretty straightforward.  It uses one routine to see if my push-to-walk button was pressed.  And another to reset the "facts" and run the rules engine every couple seconds.

# Notes
What I didn't expect was how hard it was to make a set of rules that worked the way I wanted.   Writing the rules wasn't hard, but
thinking through all the logic was.

