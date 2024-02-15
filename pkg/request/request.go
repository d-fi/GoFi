package request

func GetTrackInfo(sngID string) (map[string]interface{}, error) {
	body := map[string]interface{}{"sng_id": sngID}
	return Request(body, "song.getData")
}
