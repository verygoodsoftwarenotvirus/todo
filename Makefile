define MAGE_INQUIRY
_____________________________________________
|        Did you mean to invoke mage?       |
---------------------------------------------
            \   ^__^
             \  (oo)\_______
                (__)\       )\/\\
                    ||---W-|
                    ||    ||
endef
export MAGE_INQUIRY

define MAGE_REMINDER
_____________________________________________
| Try to remember to use Mage next time! :) |
---------------------------------------------
                         ^__^   /
                 _______\(oo)  /
            \\/\(       \(__)
                 |-W---||
                 ||    ||
endef
export MAGE_REMINDER

.DEFAULT:
	@echo "$$MAGE_INQUIRY"
	@sleep 5
	@mage $@
	@echo "$$MAGE_REMINDER"
