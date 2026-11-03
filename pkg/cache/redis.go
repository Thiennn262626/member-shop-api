package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"member-shop-api/pkg/logger"
	"reflect"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type redisCacheService struct {
	rdb    *redis.Client
	logger *zerolog.Logger
}

func (cs *redisCacheService) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	traceID := logger.GetTraceID(ctx)
	method := "GetTTL"
	//

	ttl, err := cs.rdb.TTL(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		//data not found
		cs.logger.Warn().Str("trace_id", traceID).Str("method", method).Str("key", key).Err(err).Msg("Data not found!")
		return 0, err
	}

	if err != nil {
		cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("key", key).Err(err).Msg("")
		return 0, err
	}

	// nếu không có lỗi thì ghi log info lại để theo dõi
	cs.logger.Info().Str("trace_id", traceID).Str("method", method).Str("key", key).Str("value", ttl.String()).Msg("")

	return ttl, nil
}

func (cs *redisCacheService) GetClient() *redis.Client {
	return cs.rdb
}

func NewRedisCacheService(rdb *redis.Client, logger *zerolog.Logger) RedisCacheService {
	l := logger.With().Str("module", "redis").Logger()
	return &redisCacheService{
		rdb:    rdb,
		logger: &l,
	}
}

func (cs *redisCacheService) Get(ctx context.Context, key string, dest any) error {

	traceID := logger.GetTraceID(ctx)
	method := "Get"
	//

	data, err := cs.rdb.Get(ctx, key).Result()

	if errors.Is(err, redis.Nil) {
		// Cache miss - log debug
		cs.logger.Debug().Str("trace_id", traceID).Str("method", method).Str("key", key).Msg("Cache miss")
		return err
	}

	if err != nil {
		cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("key", key).Err(err).Msg("")
		return err
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("key", key).Err(err).Msg("")
		return err
	}

	// nếu không có lỗi thì ghi log info lại để theo dõi
	cs.logger.Info().Str("trace_id", traceID).Str("method", method).Str("key", key).Int("bytes_len", len(data)).Msg("")

	return nil
}

func (cs *redisCacheService) Set(ctx context.Context, key string, value any, ttl time.Duration) error {

	traceID := logger.GetTraceID(ctx)
	method := "Set"
	//

	data, errMarshal := json.Marshal(value)
	if errMarshal != nil {
		cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("key", key).Interface("value", value).Err(errMarshal).Msg("")
		return errMarshal
	}

	if err := cs.rdb.Set(ctx, key, data, ttl).Err(); err != nil {
		cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("key", key).Int("bytes_len", len(data)).Err(err).Msg("")
		return err
	}

	// nếu không có lỗi thì ghi log info lại để theo dõi
	cs.logger.Info().Str("trace_id", traceID).Str("method", method).Str("key", key).Int("bytes_len", len(data)).Msg("")

	return nil
}

// Del implements RedisCacheService.
func (cs *redisCacheService) Del(ctx context.Context, key string) error {
	// Thực hiện xóa key trong Redis và lấy kết quả trả về
	err := cs.rdb.Del(ctx, key).Err()
	if err != nil {
		// Trả về lỗi nếu quá trình xóa thất bại
		return fmt.Errorf("failed to delete key %s from redis: %w", key, err)
	}

	return nil
}

func (cs *redisCacheService) Clear(ctx context.Context, pattern string) error {
	traceID := logger.GetTraceID(ctx)
	method := "Clear"

	cursor := uint64(0)
	for {
		keys, nextCursor, err := cs.rdb.Scan(ctx, cursor, pattern, 250).Result()
		if err != nil {
			cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("pattern", pattern).Err(err).Msg("")
			return err
		}

		if len(keys) > 0 {
			cs.rdb.Del(ctx, keys...)
		}

		cursor = nextCursor

		if cursor == 0 {
			break
		}
	}

	// nếu không có lỗi thì ghi log info lại để theo dõi
	cs.logger.Info().Str("trace_id", traceID).Str("method", method).Str("pattern", pattern).Msg("")

	return nil
}

func (cs *redisCacheService) Exists(ctx context.Context, key string) (bool, error) {

	traceID := logger.GetTraceID(ctx)
	method := "Exists"

	//
	count, err := cs.rdb.Exists(ctx, key).Result()
	if err != nil {
		cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("key", key).Err(err).Msg("")
		return false, err
	}

	// nếu không có lỗi thì ghi log info lại để theo dõi
	cs.logger.Info().Str("trace_id", traceID).Str("method", method).Str("key", key).Msg("")

	return count > 0, nil
}

func (cs *redisCacheService) Increment(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error) {
	traceID := logger.GetTraceID(ctx)
	method := "Increment"

	// 1. Thực hiện tăng giá trị
	result, err := cs.rdb.IncrBy(ctx, key, delta).Result()
	if err != nil {
		cs.logger.Error().
			Str("trace_id", traceID).
			Str("method", method).
			Str("key", key).
			Int64("delta", delta).
			Err(err).
			Msg("")

		return 0, err
	}

	// 2. Tối ưu: Chỉ set TTL nếu key vừa được tạo mới từ số 0
	// (Ví dụ: giá trị cũ không tồn tại -> Redis coi là 0 -> cộng delta -> result đúng bằng delta)
	if result == delta {
		cs.rdb.Expire(ctx, key, ttl)
	}

	// 3. Log info khi thành công
	cs.logger.Info().
		Str("trace_id", traceID).
		Str("method", method).
		Str("key", key).
		Int64("delta", delta).
		Int64("new_value", result).
		Msg("")

	return result, nil
}

func (cs *redisCacheService) ZIncrement(ctx context.Context, key string, member string, score int64, ttl time.Duration) (float64, error) {
	traceID := logger.GetTraceID(ctx)
	method := "ZIncrement"

	// Chuyển score (int64) sang float64 vì lệnh ZIncrBy của Redis làm việc với kiểu float
	scoreFloat := float64(score)

	// Tạo một pipeline để gộp lệnh, giảm RTT (Round Trip Time) tới Redis
	pipe := cs.rdb.Pipeline()

	// Lệnh 1: Tăng điểm số của member trong Sorted Set
	incrCmd := pipe.ZIncrBy(ctx, key, scoreFloat, member)

	// Lệnh 2: Đặt TTL cho key.
	// Nếu key đã có TTL, lệnh này sẽ cập nhật lại (refresh) hoặc giữ nguyên tùy bài toán của bạn.
	// Nếu bạn muốn CHỈ set TTL khi key mới tạo (chưa có TTL), xem lưu ý phía dưới.
	pipe.Expire(ctx, key, ttl)

	// Thực thi cả 2 lệnh cùng lúc
	_, err := pipe.Exec(ctx)
	if err != nil {
		// Đã sửa lỗi biến 'delta' thành 'score'
		cs.logger.Error().
			Str("trace_id", traceID).
			Str("method", method).
			Str("key", key).
			Int64("score", score).
			Err(err).
			Msg("")

		return 0, err
	}

	// Lấy kết quả trả về từ lệnh ZIncrBy
	result, err := incrCmd.Result()
	if err != nil {
		return 0, err
	}

	// Log info khi thành công
	cs.logger.Info().
		Str("trace_id", traceID).
		Str("method", method).
		Str("key", key).
		Int64("score", score).
		Float64("new_score", result).
		Msg("")

	return result, nil
}

type ZMemberWithScore struct {
	Member string  `json:"member"`
	Score  float64 `json:"score"`
}

func (cs *redisCacheService) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZMemberWithScore, error) {
	traceID := logger.GetTraceID(ctx)
	method := "ZRevRangeWithScores"

	// 1. Gọi lệnh lấy dữ liệu từ Redis (Xếp hạng từ cao xuống thấp)
	res, err := cs.rdb.ZRevRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		// Nếu không tìm thấy key (Redis trả về mảng rỗng chứ không trả về lỗi redis.Nil cho lệnh Range)
		cs.logger.Error().
			Str("trace_id", traceID).
			Str("method", method).
			Str("key", key).
			Int64("start", start).
			Int64("stop", stop).
			Err(err).
			Msg("")
		return nil, err
	}

	// 2. Chuyển đổi dữ liệu từ kiểu của go-redis sang kiểu dữ liệu tường minh của bạn
	// Khởi tạo mảng với dung lượng (capacity) bằng đúng số phần tử nhận được để tối ưu bộ nhớ
	members := make([]ZMemberWithScore, 0, len(res))
	for _, z := range res {
		// go-redis trả về z.Member dạng any (hoặc string tùy phiên bản v8/v9), ép kiểu sang string nếu cần
		memberStr, ok := z.Member.(string)
		if !ok {
			// Nếu z.Member đã là string sẵn (trong go-redis v9)
			memberStr = fmt.Sprintf("%v", z.Member)
		}

		members = append(members, ZMemberWithScore{
			Member: memberStr,
			Score:  z.Score,
		})
	}

	// 3. Ghi log info thành công (Chỉ ghi nhận số lượng phần tử trả về thay vì ghi toàn bộ data để bảo mật)
	cs.logger.Info().
		Str("trace_id", traceID).
		Str("method", method).
		Str("key", key).
		Int("result_count", len(members)).
		Msg("")

	return members, nil
}

func (cs *redisCacheService) Scan(ctx context.Context, matchPattern string) ([]string, error) {
	traceID := logger.GetTraceID(ctx)
	method := "Scan"

	// 1. Khởi tạo một Iterator để quét các key khớp với pattern (ví dụ: "cooling:*")
	// Khởi tạo Count = 100 giúp lấy 100 key mỗi đợt quét để tối ưu hiệu năng mạng
	iter := cs.rdb.Scan(ctx, 0, matchPattern, 100).Iterator()

	var keys []string

	// 2. Vòng lặp lấy toàn bộ key
	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	// 3. Kiểm tra xem quá trình quét của Iterator có gặp lỗi mạng/kết nối không
	if err := iter.Err(); err != nil {
		cs.logger.Error().
			Str("trace_id", traceID).
			Str("method", method).
			Str("pattern", matchPattern).
			Err(err).
			Msg("failed to scan redis keys")
		return nil, err
	}

	// 4. Ghi log thành công (chỉ ghi nhận số lượng key tìm thấy để bảo mật và tối ưu ổ đĩa)
	cs.logger.Info().
		Str("trace_id", traceID).
		Str("method", method).
		Str("pattern", matchPattern).
		Int("found_keys_count", len(keys)).
		Msg("redis scan success")

	return keys, nil
}

// Rename implements RedisCacheService.
func (cs *redisCacheService) Rename(ctx context.Context, tmpKey string, key string) error {
	traceID := logger.GetTraceID(ctx)
	method := "Rename"

	// Đổi tên key trong Redis
	err := cs.rdb.Rename(ctx, tmpKey, key).Err()

	if err != nil {
		cs.logger.Error().
			Str("trace_id", traceID).
			Str("method", method).
			Str("tmpKey", tmpKey).
			Str("key", key).
			Err(err).
			Msg("failed to Rename redis keys")

		return fmt.Errorf("failed to rename key from %s to %s: %w", tmpKey, key, err)
	}

	cs.logger.Info().
		Str("trace_id", traceID).
		Str("method", method).
		Str("tmpKey", tmpKey).
		Str("key", key).
		Msg("redis Rename success")
	return nil
}

// ZAdd implements RedisCacheService.
func (cs *redisCacheService) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	traceID := logger.GetTraceID(ctx)
	method := "ZAdd"

	// Thêm các member vào Sorted Set với score tương ứng
	err := cs.rdb.ZAdd(ctx, key, members...).Err()

	if err != nil {
		cs.logger.Error().
			Str("trace_id", traceID).
			Str("method", method).
			Str("key", key).
			Interface("members", members).
			Err(err).
			Msg("failed to ZAdd")

		log.Println("ZAdd member err: ", members)

		return fmt.Errorf("failed to ZAdd to key %s: %w", key, err)
	}

	cs.logger.Info().
		Str("trace_id", traceID).
		Str("method", method).
		Str("key", key).
		Interface("members", members).
		Msg("redis ZAdd success")

	return nil
}

// 1. HSET - Lưu 1 field vào Hash
func (cs *redisCacheService) HSet(ctx context.Context, hashKey string, key string, value any) error {
	traceID := logger.GetTraceID(ctx)
	method := "HSet"

	data, errMarshal := json.Marshal(value)
	if errMarshal != nil {
		cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("hashKey", hashKey).Str("key", key).Err(errMarshal).Msg("Mã hóa JSON thất bại")
		return errMarshal
	}

	if err := cs.rdb.HSet(ctx, hashKey, key, data).Err(); err != nil {
		cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("hashKey", hashKey).Str("key", key).Int("bytes_len", len(data)).Err(err).Msg("Ghi vào Redis thất bại")
		return err
	}

	cs.logger.Info().Str("trace_id", traceID).Str("method", method).Str("hashKey", hashKey).Str("key", key).Int("bytes_len", len(data)).Msg("Thành công")
	return nil
}

// 2. HGET - Lấy 1 field từ Hash và gán vào biến dest
func (cs *redisCacheService) HGet(ctx context.Context, hashKey string, key string, dest any) error {
	traceID := logger.GetTraceID(ctx)
	method := "HGet"

	val, err := cs.rdb.HGet(ctx, hashKey, key).Result()
	if err == redis.Nil {
		return err // Trả về redis.Nil để tầng service ngoài biết dữ liệu không tồn tại
	} else if err != nil {
		cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("hashKey", hashKey).Str("key", key).Err(err).Msg("Đọc từ Redis thất bại")
		return err
	}

	if errUnmarshal := json.Unmarshal([]byte(val), dest); errUnmarshal != nil {
		cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("hashKey", hashKey).Str("key", key).Err(errUnmarshal).Msg("Giải mã JSON thất bại")
		return errUnmarshal
	}

	return nil
}

// 3. HMSET - Lưu NHIỀU field cùng lúc vào Hash (Dùng map[string]any)
func (cs *redisCacheService) HMSet(ctx context.Context, hashKey string, values map[string]any) error {
	traceID := logger.GetTraceID(ctx)
	method := "HMSet"

	if len(values) == 0 {
		return nil
	}

	// Chuyển đổi map[string]any thành map[string]interface{} chứa JSON string để gửi lên Redis
	bulkData := make(map[string]interface{})
	for k, v := range values {
		data, errMarshal := json.Marshal(v)
		if errMarshal != nil {
			cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("hashKey", hashKey).Str("key", k).Err(errMarshal).Msg("Mã hóa JSON thất bại trong chuỗi Bulk")
			return errMarshal
		}
		bulkData[k] = string(data)
	}

	// Hàm HSet của go-redis hỗ trợ truyền map để ghi bulk một cách tối ưu
	if err := cs.rdb.HMSet(ctx, hashKey, bulkData).Err(); err != nil {
		cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("hashKey", hashKey).Int("items_count", len(bulkData)).Err(err).Msg("Ghi Bulk vào Redis thất bại")
		return err
	}

	cs.logger.Info().Str("trace_id", traceID).Str("method", method).Str("hashKey", hashKey).Int("items_count", len(bulkData)).Msg("Thành công")
	return nil
}

// 4. HMGET - Lấy NHIỀU field cùng lúc và tự động parse thành Slice
func (cs *redisCacheService) HMGet(ctx context.Context, hashKey string, keys []string, dest any) error {
	traceID := logger.GetTraceID(ctx)
	method := "HMGet"

	if len(keys) == 0 {
		return nil
	}

	// Kiểm tra xem biến dest truyền vào có phải là con trỏ trỏ tới 1 Slice không (Bắt buộc để Unmarshal danh sách)
	destVal := reflect.ValueOf(dest)
	if destVal.Kind() != reflect.Ptr || destVal.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("dest phải là một con trỏ trỏ tới một slice (ví dụ: *[]MyStruct)")
	}

	sliceVal := destVal.Elem()
	elemType := sliceVal.Type().Elem() // Lấy kiểu dữ liệu của phần tử trong Slice (ví dụ: MyStruct)

	// Gọi lệnh HMGet lên Redis
	results, err := cs.rdb.HMGet(ctx, hashKey, keys...).Result()
	if err != nil {
		cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("hashKey", hashKey).Err(err).Msg("Đọc Bulk từ Redis thất bại")
		return err
	}

	// Tạo một slice tạm thời để chứa dữ liệu sau khi parse JSON
	tempSlice := reflect.MakeSlice(sliceVal.Type(), len(results), len(results))

	for i, val := range results {
		// Tạo một phần tử rỗng mới cùng kiểu với phần tử của slice kết quả
		newElemPtr := reflect.New(elemType)

		if val != nil {
			// Ép kiểu interface từ go-redis về string
			jsonStr, ok := val.(string)
			if ok {
				// Parse JSON trực tiếp vào con trỏ của phần tử mới tạo
				if errUnmarshal := json.Unmarshal([]byte(jsonStr), newElemPtr.Interface()); errUnmarshal != nil {
					cs.logger.Error().Str("trace_id", traceID).Str("method", method).Str("hashKey", hashKey).Str("key", keys[i]).Err(errUnmarshal).Msg("Giải mã JSON thất bại trong chuỗi Bulk")
					return errUnmarshal
				}
			}
		}
		// Gán phần tử (giá trị thực, không phải con trỏ) vào slice tạm
		tempSlice.Index(i).Set(newElemPtr.Elem())
	}

	// Gán ngược slice tạm vào biến dest mà user truyền vào ban đầu
	sliceVal.Set(tempSlice)

	cs.logger.Info().Str("trace_id", traceID).Str("method", method).Str("hashKey", hashKey).Int("req_keys", len(keys)).Msg("Thành công")
	return nil
}
