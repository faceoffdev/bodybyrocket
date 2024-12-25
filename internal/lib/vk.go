package lib

import (
	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/SevereCloud/vksdk/v3/object"
	"iter"
)

func IterateWallPosts(vk *api.VK, groupId int) iter.Seq2[*object.WallWallpost, error] {
	return func(yield func(*object.WallWallpost, error) bool) {
		const maxCountItems = 10

		var (
			offset int
			wall   api.WallGetResponse
			err    error
		)

		for ok := true; ok; ok = offset < wall.Count {
			wall, err = vk.WallGet(api.Params{"owner_id": groupId, "count": maxCountItems, "offset": offset})
			if err != nil {
				yield(nil, err)
				return
			}

			for _, post := range wall.Items {
				if !yield(&post, nil) {
					return
				}
			}

			offset += maxCountItems
		}
	}
}
