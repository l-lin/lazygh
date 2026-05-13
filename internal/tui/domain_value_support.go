package tui

import (
	"encoding/json"
	"reflect"

	githubdomain "github.com/l-lin/lazygh/internal/github"
)

func toDomainConnectedUser(user any) (githubdomain.ConnectedUser, bool) {
	return decodeDomainValue[githubdomain.ConnectedUser](user)
}

func toDomainNotification(notification any) (githubdomain.Notification, bool) {
	return decodeDomainValue[githubdomain.Notification](notification)
}

func toDomainPullRequestSummary(summary any) (githubdomain.PullRequest, bool) {
	return decodeDomainValue[githubdomain.PullRequest](summary)
}

func toDomainPullRequestDetail(detail any) (githubdomain.PullRequestDetail, bool) {
	return decodeDomainValue[githubdomain.PullRequestDetail](detail)
}

func toDomainRepository(repository any) (githubdomain.Repository, bool) {
	return decodeDomainValue[githubdomain.Repository](repository)
}

func toDomainPullRequestInlineComment(comment any) (githubdomain.PullRequestInlineComment, bool) {
	return decodeDomainValue[githubdomain.PullRequestInlineComment](comment)
}

func toDomainPullRequestDiff(diff any) (githubdomain.PullRequestDiff, bool) {
	return decodeDomainValue[githubdomain.PullRequestDiff](diff)
}

func toDomainPullRequestComment(comment any) (githubdomain.PullRequestComment, bool) {
	return decodeDomainValue[githubdomain.PullRequestComment](comment)
}

func toDomainReviewThreadTarget(target any) (githubdomain.PullRequestReviewThreadTarget, bool) {
	return decodeDomainValue[githubdomain.PullRequestReviewThreadTarget](target)
}

func toDomainPullRequestStatusCheck(check any) (githubdomain.PullRequestStatusCheck, bool) {
	return decodeDomainValue[githubdomain.PullRequestStatusCheck](check)
}

func toDomainPullRequestReviewThread(thread any) (githubdomain.PullRequestReviewThread, bool) {
	return decodeDomainValue[githubdomain.PullRequestReviewThread](thread)
}

func toDomainPullRequestComments(comments any) []githubdomain.PullRequestComment {
	actual, _ := decodeDomainValue[[]githubdomain.PullRequestComment](comments)
	return actual
}

func toDomainPullRequestReviewThreads(threads any) []githubdomain.PullRequestReviewThread {
	actual, _ := decodeDomainValue[[]githubdomain.PullRequestReviewThread](threads)
	return actual
}

func toDomainPullRequestInlineComments(comments any) []githubdomain.PullRequestInlineComment {
	actual, _ := decodeDomainValue[[]githubdomain.PullRequestInlineComment](comments)
	return actual
}

func toDomainPullRequestReviews(reviews any) []githubdomain.PullRequestReview {
	actual, _ := decodeDomainValue[[]githubdomain.PullRequestReview](reviews)
	return actual
}

func toDomainPullRequestCommits(commits any) []githubdomain.PullRequestCommit {
	actual, _ := decodeDomainValue[[]githubdomain.PullRequestCommit](commits)
	return actual
}

func toDomainReactionGroup(group any) (githubdomain.ReactionGroup, bool) {
	return decodeDomainValue[githubdomain.ReactionGroup](group)
}

func toDomainReactionGroups(groups any) []githubdomain.ReactionGroup {
	actual, _ := decodeDomainValue[[]githubdomain.ReactionGroup](groups)
	return actual
}

func decodeDomainValue[T any](source any) (T, bool) {
	var target T
	if source == nil {
		return target, false
	}

	targetValue := reflect.ValueOf(&target).Elem()
	if copyCompatibleValue(targetValue, reflect.ValueOf(source)) {
		return target, true
	}

	payload, err := json.Marshal(source)
	if err != nil {
		return target, false
	}
	if err := json.Unmarshal(payload, &target); err != nil {
		return target, false
	}
	return target, true
}

func copyCompatibleValue(target reflect.Value, source reflect.Value) bool {
	if !target.IsValid() || !target.CanSet() || !source.IsValid() {
		return false
	}

	if source.Kind() == reflect.Interface && !source.IsNil() {
		source = source.Elem()
	}
	if target.Kind() == reflect.Interface {
		if source.Type().AssignableTo(target.Type()) {
			target.Set(source)
			return true
		}
		return false
	}
	if source.Type().AssignableTo(target.Type()) {
		target.Set(source)
		return true
	}
	if source.Type().ConvertibleTo(target.Type()) && isScalarKind(target.Kind()) {
		target.Set(source.Convert(target.Type()))
		return true
	}

	switch target.Kind() {
	case reflect.Pointer:
		if source.Kind() == reflect.Pointer {
			if source.IsNil() {
				return false
			}
			return copyCompatibleValue(target, source.Elem())
		}
		allocated := reflect.New(target.Type().Elem())
		if !copyCompatibleValue(allocated.Elem(), source) {
			return false
		}
		target.Set(allocated)
		return true
	case reflect.Struct:
		if source.Kind() == reflect.Pointer {
			if source.IsNil() {
				return false
			}
			source = source.Elem()
		}
		if source.Kind() != reflect.Struct {
			return false
		}
		copiedAny := false
		for index := 0; index < target.NumField(); index++ {
			targetField := target.Field(index)
			if !targetField.CanSet() {
				continue
			}
			targetStructField := target.Type().Field(index)
			sourceField := source.FieldByName(targetStructField.Name)
			if !sourceField.IsValid() {
				continue
			}
			if copyCompatibleValue(targetField, sourceField) {
				copiedAny = true
			}
		}
		return copiedAny
	case reflect.Slice:
		if source.Kind() == reflect.Pointer {
			if source.IsNil() {
				return false
			}
			source = source.Elem()
		}
		if source.Kind() != reflect.Slice {
			return false
		}
		copied := reflect.MakeSlice(target.Type(), 0, source.Len())
		for index := 0; index < source.Len(); index++ {
			targetElement := reflect.New(target.Type().Elem()).Elem()
			if copyCompatibleValue(targetElement, source.Index(index)) {
				copied = reflect.Append(copied, targetElement)
				continue
			}
			return false
		}
		target.Set(copied)
		return true
	case reflect.Map:
		if source.Kind() == reflect.Pointer {
			if source.IsNil() {
				return false
			}
			source = source.Elem()
		}
		if source.Kind() != reflect.Map {
			return false
		}
		copied := reflect.MakeMapWithSize(target.Type(), source.Len())
		iterator := source.MapRange()
		for iterator.Next() {
			targetKey := reflect.New(target.Type().Key()).Elem()
			if !copyCompatibleValue(targetKey, iterator.Key()) {
				return false
			}
			targetValue := reflect.New(target.Type().Elem()).Elem()
			if !copyCompatibleValue(targetValue, iterator.Value()) {
				return false
			}
			copied.SetMapIndex(targetKey, targetValue)
		}
		target.Set(copied)
		return true
	default:
		return false
	}
}

func isScalarKind(kind reflect.Kind) bool {
	switch kind {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.String:
		return true
	default:
		return false
	}
}
