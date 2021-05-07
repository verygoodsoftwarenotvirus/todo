define MAGE_INQUIRY
_______________________________________________
|         Did you mean to invoke mage?        |
-----------------------------------------------
             \   ^__^
              \  (oo)\_______
                 (__)\       )\/\\
                     ||---W-|
                     ||    ||
endef
export MAGE_INQUIRY

define MAGE_REMINDER
_______________________________________________
| Try to remember to use `mage` next time! :) |
-----------------------------------------------
                         ^__^   /
                 _______/(oo)  /
            //\/(       /(__)
                 |-W---||
                 ||    ||
endef
export MAGE_REMINDER

.DEFAULT:
	@echo "$$MAGE_INQUIRY"
	@sleep 5
	@mage $@
	@echo "$$MAGE_REMINDER"
