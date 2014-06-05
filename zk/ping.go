package zk

type pingRequest struct {
	xid    int32
	opcode int32
}

func encodePingRequest(buf []byte, req *pingRequest) int32 {
	Int32ToBytes(buf[4:], req.xid)
	Int32ToBytes(buf[8:], req.opcode)
	Int32ToBytes(buf[0:], 8)
	return 12
}

type pingResponse struct {
}

func decodePingResponse(buf []byte, res *pingResponse) {

}
