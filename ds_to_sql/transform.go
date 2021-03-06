package ds_to_sql

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/golang/protobuf/proto"
	"github.com/syndtr/goleveldb/leveldb/journal"
	pb "github.com/sonnes/loki/ds_to_sql/pb"
)

type CSVMixin interface {
	ExtractCSV() *[][]string
	SetKey(*Key)
}

func ReadDSFile(filePath string, newDst func() CSVMixin) (*[][]string, error) {

	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	journals := journal.NewReader(f, nil, false, true)

	allCSVRows := make([][]string, 0)

	for {
		j, err := journals.Next()

		if err != nil {
			break
		}

		b, err := ioutil.ReadAll(j)

		if err != nil {
			break
		}

		pb := &pb.EntityProto{}
		if err := proto.Unmarshal(b, pb); err != nil {
			log.Fatal(err)
			break
		}

		dst := newDst()

		// protobuf to entity
		err = LoadEntity(dst, pb)

		if err != nil {
			log.Fatal(err)
		}

		key, _ := protoToKey(pb.GetKey())
		dst.SetKey(key)

		// entity to csv
		entityRows := *dst.ExtractCSV()

		allCSVRows = append(allCSVRows, entityRows...)
	}

	return &allCSVRows, nil
}
