define MAGE_INQUIRY
________________________________
| Did you mean to invoke mage? |
--------------------------------
      \   ^__^
       \  (oo)\_______
          (__)\       )\/\\
              ||----w |
              ||     ||
endef

export MAGE_INQUIRY
.DEFAULT:
	@echo "$$MAGE_INQUIRY"
	@mage $@
