package frontend2

import "net/http"

const svgFaviconSrc = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 512 512">
<style>
  path {
    fill: #666;
  }

  @media (prefers-color-scheme: dark) {
    path {
      fill: #FFFFFF;
    } 
  }
</style>
<!-- actual SVG content goes here -->
</svg>`

func (s *Service) favicon(res http.ResponseWriter, _ *http.Request) {
	if _, err := res.Write([]byte(svgFaviconSrc)); err != nil {
		s.panicker.Panic(err)
	}
}
