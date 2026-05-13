package githubcli

type NotificationAssembler struct{}

type notificationDTO Notification

type issueDetailDTO IssueDetail

type releaseDetailDTO ReleaseDetail

func (NotificationAssembler) ParseList(stdout []byte) ([]Notification, error) {
	var dtos []notificationDTO
	if err := decodeEndpointPaginatedOrFlatJSONResponse(stdout, &dtos, ErrInvalidNotificationResponse); err != nil {
		return nil, err
	}
	return mapNotificationDTOs(dtos), nil
}

func (NotificationAssembler) ParseIssueDetail(stdout []byte) (IssueDetail, error) {
	var dto issueDetailDTO
	if err := decodeEndpointJSONResponse(stdout, &dto, ErrInvalidIssueDetailResponse); err != nil {
		return IssueDetail{}, err
	}
	return IssueDetail(dto).normalized(), nil
}

func (NotificationAssembler) ParseReleaseDetail(stdout []byte) (ReleaseDetail, error) {
	var dto releaseDetailDTO
	if err := decodeEndpointJSONResponse(stdout, &dto, ErrInvalidReleaseDetailResponse); err != nil {
		return ReleaseDetail{}, err
	}
	return ReleaseDetail(dto).normalized(), nil
}

func mapNotificationDTOs(dtos []notificationDTO) []Notification {
	if len(dtos) == 0 {
		return nil
	}

	notifications := make([]Notification, 0, len(dtos))
	for _, dto := range dtos {
		notification := Notification(dto).normalized()
		if notification.Done {
			continue
		}
		notifications = append(notifications, notification)
	}
	if len(notifications) == 0 {
		return nil
	}
	return notifications
}
