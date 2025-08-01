package scalingo

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type Event struct {
	ID          string                 `json:"id"`
	AppID       string                 `json:"app_id"`
	CreatedAt   time.Time              `json:"created_at"`
	User        EventUser              `json:"user"`
	Type        EventTypeName          `json:"type"`
	AppName     string                 `json:"app_name"`
	RawTypeData json.RawMessage        `json:"type_data"`
	TypeData    map[string]interface{} `json:"-"`
	ProjectID   string                 `json:"project_id"`
	ProjectName string                 `json:"project_name"`
}

type EventSecurityTypeData struct {
	RemoteIP string `json:"remote_ip"`
}

func (ev *Event) GetEvent() *Event {
	return ev
}

func (ev *Event) TypeDataPtr() interface{} {
	return ev.TypeData
}

func (ev *Event) String() string {
	return fmt.Sprintf("Unknown event %v on app %v", ev.Type, ev.AppName)
}

func (ev *Event) When() string {
	return ev.CreatedAt.Format("Mon Jan 02 2006 15:04:05")
}

func (ev *Event) Who() string {
	return fmt.Sprintf("%s (%s)", ev.User.Username, ev.User.Email)
}

func (ev *Event) PrintableType() string {
	typeName := strings.ReplaceAll(string(ev.Type), "_", " ")
	return cases.Title(language.English).String(typeName)
}

type DetailedEvent interface {
	fmt.Stringer
	GetEvent() *Event
	PrintableType() string
	When() string
	Who() string
	TypeDataPtr() interface{}
}

type Events []DetailedEvent

type EventUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	ID       string `json:"id"`
}

type EventTypeName string

const (
	EventNewUser                     EventTypeName = "new_user"
	EventNewApp                      EventTypeName = "new_app"
	EventEditApp                     EventTypeName = "edit_app"
	EventDeleteApp                   EventTypeName = "delete_app"
	EventRenameApp                   EventTypeName = "rename_app"
	EventUpdateAppProject            EventTypeName = "update_app_project"
	EventTransferApp                 EventTypeName = "transfer_app"
	EventRestart                     EventTypeName = "restart"
	EventScale                       EventTypeName = "scale"
	EventStopApp                     EventTypeName = "stop_app"
	EventCrash                       EventTypeName = "crash"
	EventRepeatedCrash               EventTypeName = "repeated_crash"
	EventDeployment                  EventTypeName = "deployment"
	EventLinkSCM                     EventTypeName = "link_scm"
	EventUpdateSCM                   EventTypeName = "update_scm"
	EventUnlinkSCM                   EventTypeName = "unlink_scm"
	EventNewIntegration              EventTypeName = "new_integration"
	EventDeleteIntegration           EventTypeName = "delete_integration"
	EventAuthorizeGithub             EventTypeName = "authorize_github"
	EventRevokeGithub                EventTypeName = "revoke_github"
	EventRun                         EventTypeName = "run"
	EventNewDomain                   EventTypeName = "new_domain"
	EventEditDomain                  EventTypeName = "edit_domain"
	EventDeleteDomain                EventTypeName = "delete_domain"
	EventUpgradeDatabase             EventTypeName = "upgrade_database"
	EventNewAddon                    EventTypeName = "new_addon"
	EventUpgradeAddon                EventTypeName = "upgrade_addon"
	EventDeleteAddon                 EventTypeName = "delete_addon"
	EventResumeAddon                 EventTypeName = "resume_addon"
	EventSuspendAddon                EventTypeName = "suspend_addon"
	EventDatabaseAddFeature          EventTypeName = "database/add_feature"
	EventDatabaseRemoveFeature       EventTypeName = "database/remove_feature"
	EventNewCollaborator             EventTypeName = "new_collaborator"
	EventAcceptCollaborator          EventTypeName = "accept_collaborator"
	EventDeleteCollaborator          EventTypeName = "delete_collaborator"
	EventChangeCollaboratorRole      EventTypeName = "change_collaborator_role"
	EventNewVariable                 EventTypeName = "new_variable"
	EventEditVariable                EventTypeName = "edit_variable"
	EventEditVariables               EventTypeName = "edit_variables"
	EventDeleteVariable              EventTypeName = "delete_variable"
	EventAddCredit                   EventTypeName = "add_credit"
	EventAddPaymentMethod            EventTypeName = "add_payment_method"
	EventAddVoucher                  EventTypeName = "add_voucher"
	EventNewKey                      EventTypeName = "new_key"
	EventEditKey                     EventTypeName = "edit_key"
	EventDeleteKey                   EventTypeName = "delete_key"
	EventPaymentAttempt              EventTypeName = "payment_attempt"
	EventNewAlert                    EventTypeName = "new_alert"
	EventAlert                       EventTypeName = "alert"
	EventDeleteAlert                 EventTypeName = "delete_alert"
	EventNewAutoscaler               EventTypeName = "new_autoscaler"
	EventEditAutoscaler              EventTypeName = "edit_autoscaler"
	EventDeleteAutoscaler            EventTypeName = "delete_autoscaler"
	EventAddonUpdated                EventTypeName = "addon_updated"
	EventStartRegionMigration        EventTypeName = "start_region_migration"
	EventNewLogDrain                 EventTypeName = "new_log_drain"
	EventDeleteLogDrain              EventTypeName = "delete_log_drain"
	EventNewAddonLogDrain            EventTypeName = "new_addon_log_drain"
	EventDeleteAddonLogDrain         EventTypeName = "delete_addon_log_drain"
	EventNewNotifier                 EventTypeName = "new_notifier"
	EventEditNotifier                EventTypeName = "edit_notifier"
	EventDeleteNotifier              EventTypeName = "delete_notifier"
	EventEditHDSContact              EventTypeName = "edit_hds_contact"
	EventCreateDataAccessConsent     EventTypeName = "create_data_access_consent"
	EventNewToken                    EventTypeName = "new_token"
	EventRegenerateToken             EventTypeName = "regenerate_token"
	EventDeleteToken                 EventTypeName = "delete_token"
	EventTfaEnabled                  EventTypeName = "tfa_enabled"
	EventTfaDisabled                 EventTypeName = "tfa_disabled"
	EventLoginSuccess                EventTypeName = "login_success"
	EventLoginFailure                EventTypeName = "login_failure"
	EventLoginLock                   EventTypeName = "login_lock"
	EventLoginUnlockSuccess          EventTypeName = "login_unlock_success"
	EventPasswordResetQuery          EventTypeName = "password_reset_query"
	EventPasswordResetSuccess        EventTypeName = "password_reset_success"
	EventStackChanged                EventTypeName = "stack_changed"
	EventCreateReviewApp             EventTypeName = "create_review_app"
	EventDestroyReviewApp            EventTypeName = "destroy_review_app"
	EventPlanDatabaseMaintenance     EventTypeName = "plan_database_maintenance"
	EventStartDatabaseMaintenance    EventTypeName = "start_database_maintenance"
	EventCompleteDatabaseMaintenance EventTypeName = "complete_database_maintenance"

	// EventLinkGithub and EventUnlinkGithub events are kept for
	// retro-compatibility. They are replaced by SCM events.
	EventLinkGithub   EventTypeName = "link_github"
	EventUnlinkGithub EventTypeName = "unlink_github"

	// Project scoped events
	EventDeleteProject EventTypeName = "delete_project"
	EventNewProject    EventTypeName = "new_project"
	EventEditProject   EventTypeName = "edit_project"
)

type EventNewUserType struct {
	Event
	TypeData EventNewUserTypeData `json:"type_data"`
}

func (ev *EventNewUserType) String() string {
	return "You joined Scalingo. Hooray!"
}

type EventCreateReviewAppTypeData struct {
	AppID                    string `json:"app_id"`
	ReviewAppName            string `json:"review_app_name"`
	ReviewAppURL             string `json:"review_app_url"`
	SourceRepoName           string `json:"source_repo_name"`
	SourceRepoURL            string `json:"source_repo_url"`
	PullRequestName          string `json:"pr_name"`
	PullRequestNumber        int    `json:"pr_number"`
	PullRequestURL           string `json:"pr_url"`
	PullRequestComesFromFork bool   `json:"pr_comes_from_a_fork"`
}

type EventCreateReviewAppType struct {
	Event
	TypeData EventCreateReviewAppTypeData `json:"type_data"`
}

func (ev *EventCreateReviewAppType) String() string {
	return fmt.Sprintf("the review app %s has been created from the pull request %s #%d", ev.TypeData.ReviewAppName, ev.TypeData.PullRequestName, ev.TypeData.PullRequestNumber)
}

type EventDestroyReviewAppTypeData struct {
	AppID                    string `json:"app_id"`
	ReviewAppName            string `json:"review_app_name"`
	SourceRepoName           string `json:"source_repo_name"`
	SourceRepoURL            string `json:"source_repo_url"`
	PullRequestName          string `json:"pr_name"`
	PullRequestNumber        int    `json:"pr_number"`
	PullRequestURL           string `json:"pr_url"`
	PullRequestComesFromFork bool   `json:"pr_comes_from_a_fork"`
}

type EventDestroyReviewAppType struct {
	Event
	TypeData EventCreateReviewAppTypeData `json:"type_data"`
}

func (ev *EventDestroyReviewAppType) String() string {
	return fmt.Sprintf("the review app %s has been destroyed", ev.TypeData.ReviewAppName)
}

type EventNewUserTypeData struct {
}

type EventLinkGithubType struct {
	Event
	TypeData EventLinkGithubTypeData `json:"type_data"`
}

func (ev *EventLinkGithubType) String() string {
	return fmt.Sprintf("app has been linked to Github repository '%s'", ev.TypeData.RepoName)
}

type EventLinkGithubTypeData struct {
	RepoName       string `json:"repo_name"`
	LinkerUsername string `json:"linker_username"`
	GithubSource   string `json:"github_source"`
}

type EventUnlinkGithubType struct {
	Event
	TypeData EventUnlinkGithubTypeData `json:"type_data"`
}

func (ev *EventUnlinkGithubType) String() string {
	return fmt.Sprintf("app has been unlinked from Github repository '%s'", ev.TypeData.RepoName)
}

type EventUnlinkGithubTypeData struct {
	RepoName         string `json:"repo_name"`
	UnlinkerUsername string `json:"unlinker_username"`
	GithubSource     string `json:"github_source"`
}

type EventLinkSCMType struct {
	Event
	TypeData EventLinkSCMTypeData `json:"type_data"`
}

func (ev *EventLinkSCMType) String() string {
	return fmt.Sprintf("app has been linked to repository '%s'", ev.TypeData.RepoName)
}

type EventLinkSCMTypeData struct {
	RepoName                 string `json:"repo_name"`
	LinkerUsername           string `json:"linker_username"`
	Source                   string `json:"source"`
	Branch                   string `json:"branch"`
	AutoDeploy               bool   `json:"auto_deploy"`
	AutoDeployReviewApps     bool   `json:"auto_deploy_review_apps"`
	DeleteOnClose            bool   `json:"delete_on_close"`
	DeleteStale              bool   `json:"delete_stale"`
	HoursBeforeDeleteOnClose int    `json:"hours_before_delete_on_close"`
	HoursBeforeDeleteStale   int    `json:"hours_before_delete_stale"`
	CreationFromForksAllowed bool   `json:"creation_from_forks_allowed"`
}

type EventUpdateSCMType struct {
	Event
	TypeData EventLinkSCMTypeData `json:"type_data"`
}

func (ev *EventUpdateSCMType) String() string {
	return fmt.Sprintf("the link between the app and the repository '%s' has been updated", ev.TypeData.RepoName)
}

type EventUnlinkSCMType struct {
	Event
	TypeData EventUnlinkSCMTypeData `json:"type_data"`
}

func (ev *EventUnlinkSCMType) String() string {
	return fmt.Sprintf("app has been unlinked from repository '%s'", ev.TypeData.RepoName)
}

type EventUnlinkSCMTypeData struct {
	RepoName         string `json:"repo_name"`
	UnlinkerUsername string `json:"unlinker_username"`
	Source           string `json:"source"`
}

type EventRunType struct {
	Event
	TypeData EventRunTypeData `json:"type_data"`
}

func (ev *EventRunType) String() string {
	detached := ""
	if ev.TypeData.Detached {
		detached = "detached "
	}

	if ev.isEventRunFromOperator() {
		// The command executed is not available to end user if it's executed by a Scalingo operator
		return fmt.Sprintf("%sone-off container for maintenance/support purposes", detached)
	}

	return fmt.Sprintf("%sone-off container with command '%s'", detached, ev.TypeData.Command)
}

func (ev *EventRunType) Who() string {
	if ev.User.Email == "deploy@scalingo.com" {
		return "Scalingo Operator"
	}

	return ev.Event.Who()
}

func (ev *EventRunType) isEventRunFromOperator() bool {
	return ev.TypeData.Command == ""
}

type EventRunTypeData struct {
	Command    string `json:"command"`
	AuditLogID string `json:"audit_log_id"`
	Detached   bool   `json:"detached"`
}

type EventNewDomainType struct {
	Event
	TypeData EventNewDomainTypeData `json:"type_data"`
}

func (ev *EventNewDomainType) String() string {
	return fmt.Sprintf("'%s' has been associated", ev.TypeData.Hostname)
}

type EventNewDomainTypeData struct {
	Hostname string `json:"hostname"`
	SSL      bool   `json:"ssl"`
}

type EventEditDomainType struct {
	Event
	TypeData EventEditDomainTypeData `json:"type_data"`
}

func (ev *EventEditDomainType) String() string {
	t := ev.TypeData
	res := fmt.Sprintf("'%s' modified", t.Hostname)
	if !t.SSL && t.OldSSL {
		res += ", TLS certificate has been removed"
	} else if t.SSL && !t.OldSSL {
		res += ", TLS certificate has been added"
	} else if t.SSL && t.OldSSL {
		res += ", TLS certificate has been changed"
	}
	return res
}

type EventEditDomainTypeData struct {
	Hostname string `json:"hostname"`
	SSL      bool   `json:"ssl"`
	OldSSL   bool   `json:"old_ssl"`
}

type EventDeleteDomainType struct {
	Event
	TypeData EventDeleteDomainTypeData `json:"type_data"`
}

func (ev *EventDeleteDomainType) String() string {
	return fmt.Sprintf("'%s' has been disassociated", ev.TypeData.Hostname)
}

type EventDeleteDomainTypeData struct {
	Hostname string `json:"hostname"`
}

type EventCollaborator struct {
	EventUser
	Inviter   EventUser `json:"inviter"`
	IsLimited bool      `json:"is_limited"`
}

type EventNewCollaboratorType struct {
	Event
	TypeData EventNewCollaboratorTypeData `json:"type_data"`
}

func (ev *EventNewCollaboratorType) String() string {
	return fmt.Sprintf(
		"'%s' has been invited",
		ev.TypeData.Collaborator.Email,
	)
}

type EventNewCollaboratorTypeData struct {
	Collaborator EventCollaborator `json:"collaborator"`
}

type EventAcceptCollaboratorType struct {
	Event
	TypeData EventAcceptCollaboratorTypeData `json:"type_data"`
}

func (ev *EventAcceptCollaboratorType) String() string {
	return fmt.Sprintf(
		"'%s' (%s) has accepted the collaboration",
		ev.TypeData.Collaborator.Username,
		ev.TypeData.Collaborator.Email,
	)
}

// Inviter is filled there
type EventAcceptCollaboratorTypeData struct {
	Collaborator EventCollaborator `json:"collaborator"`
}

type EventDeleteCollaboratorType struct {
	Event
	TypeData EventDeleteCollaboratorTypeData `json:"type_data"`
}

func (ev *EventDeleteCollaboratorType) String() string {
	return fmt.Sprintf(
		"'%s' (%s) is not a collaborator anymore",
		ev.TypeData.Collaborator.Username,
		ev.TypeData.Collaborator.Email,
	)
}

type EventDeleteCollaboratorTypeData struct {
	Collaborator EventCollaborator `json:"collaborator"`
}

type EventChangeCollaboratorRoleType struct {
	Event
	TypeData EventChangeCollaboratorRoleTypeData `json:"type_data"`
}

type EventChangeCollaboratorRoleTypeData struct {
	Collaborator EventCollaborator `json:"collaborator"`
}

func (ev *EventChangeCollaboratorRoleType) String() string {
	role := "Collaborator"
	if ev.TypeData.Collaborator.IsLimited {
		role = "Limited collaborator"
	}

	if ev.TypeData.Collaborator.Username == "" {
		return fmt.Sprintf("%s is now a %s", ev.TypeData.Collaborator.Email, role)
	}

	return fmt.Sprintf("'%s' (%s) is now a %s", ev.TypeData.Collaborator.Username, ev.TypeData.Collaborator.Email, role)
}

type EventUpgradeDatabaseType struct {
	Event
	TypeData EventUpgradeDatabaseTypeData `json:"type_data"`
}

type EventUpgradeDatabaseTypeData struct {
	AddonName  string `json:"addon_name"`
	OldVersion string `json:"old_version"`
	NewVersion string `json:"new_version"`
}

func (ev *EventUpgradeDatabaseType) String() string {
	return fmt.Sprintf(
		"'%s' upgraded from v%s to v%s",
		ev.TypeData.AddonName, ev.TypeData.OldVersion, ev.TypeData.NewVersion,
	)
}

func (ev *EventUpgradeDatabaseType) Who() string {
	if ev.TypeData.AddonName != "" {
		return fmt.Sprintf("Addon %s", ev.TypeData.AddonName)
	}

	return ev.Event.Who()
}

type EventVariable struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type EventNewVariableType struct {
	Event
	TypeData EventNewVariableTypeData `json:"type_data"`
}

func (ev *EventNewVariableType) String() string {
	return fmt.Sprintf("'%s' added to the environment", ev.TypeData.Name)
}

func (ev *EventNewVariableType) Who() string {
	if ev.TypeData.AddonName != "" {
		return fmt.Sprintf("Addon %s", ev.TypeData.AddonName)
	}

	return ev.Event.Who()
}

type EventNewVariableTypeData struct {
	AddonName string `json:"addon_name"`
	EventVariable
}

type EventVariables []EventVariable

func (evs EventVariables) Names() string {
	names := []string{}
	for _, e := range evs {
		names = append(names, e.Name)
	}
	return strings.Join(names, ", ")
}

type EventEditVariableType struct {
	Event
	TypeData EventEditVariableTypeData `json:"type_data"`
}

func (ev *EventEditVariableType) String() string {
	return fmt.Sprintf("'%s' modified", ev.TypeData.Name)
}

type EventEditVariableTypeData struct {
	EventVariable
	OldValue  string `json:"old_value"`
	AddonName string `json:"addon_name"`
}

type EventEditVariablesType struct {
	Event
	TypeData EventEditVariablesTypeData `json:"type_data"`
}

func (ev *EventEditVariablesType) String() string {
	res := "environment changes:"
	if len(ev.TypeData.NewVars) > 0 {
		res += fmt.Sprintf(" %s added", ev.TypeData.NewVars.Names())
	}
	if len(ev.TypeData.UpdatedVars) > 0 {
		res += fmt.Sprintf(" %s modified", ev.TypeData.UpdatedVars.Names())
	}
	if len(ev.TypeData.DeletedVars) > 0 {
		res += fmt.Sprintf(" %s removed", ev.TypeData.DeletedVars.Names())
	}
	return res
}

func (ev *EventEditVariableType) Who() string {
	if ev.TypeData.AddonName != "" {
		return fmt.Sprintf("Addon %s", ev.TypeData.AddonName)
	}

	return ev.Event.Who()
}

type EventEditVariablesTypeData struct {
	NewVars     EventVariables `json:"new_vars"`
	UpdatedVars EventVariables `json:"updated_vars"`
	DeletedVars EventVariables `json:"deleted_vars"`
}

type EventDeleteVariableType struct {
	Event
	TypeData EventDeleteVariableTypeData `json:"type_data"`
}

func (ev *EventDeleteVariableType) String() string {
	return fmt.Sprintf("'%s' removed from environment", ev.TypeData.Name)
}

type EventDeleteVariableTypeData struct {
	EventVariable
}

type EventPaymentAttemptTypeData struct {
	Amount        float32 `json:"amount"`
	PaymentMethod string  `json:"payment_method"`
	Status        string  `json:"status"`
}

type EventPaymentAttemptType struct {
	Event
	TypeData EventPaymentAttemptTypeData `json:"type_data"`
}

func (ev *EventPaymentAttemptType) String() string {
	res := "Payment attempt of "
	res += fmt.Sprintf("%0.2f€", ev.TypeData.Amount)
	res += " with your "
	if ev.TypeData.PaymentMethod == "credit" {
		res += "credits"
	} else {
		res += "card"
	}
	if ev.TypeData.Status == "new" {
		res += " (pending)"
	} else if ev.TypeData.Status == "paid" {
		res += " (success)"
	} else {
		res += " (fail)"
	}
	return res
}

type EventNewAutoscalerTypeData struct {
	ContainerType string  `json:"container_type"`
	MinContainers int     `json:"min_containers,string"`
	MaxContainers int     `json:"max_containers,string"`
	Metric        string  `json:"metric"`
	Target        float64 `json:"target"`
	TargetText    string  `json:"target_text"`
}

type EventNewAutoscalerType struct {
	Event
	TypeData EventNewAutoscalerTypeData `json:"type_data"`
}

func (ev *EventNewAutoscalerType) String() string {
	d := ev.TypeData
	return fmt.Sprintf("Autoscaler created about %s on container %s (target: %s)", d.Metric, d.ContainerType, d.TargetText)
}

type EventEditAutoscalerTypeData struct {
	ContainerType string  `json:"container_type"`
	MinContainers int     `json:"min_containers,string"`
	MaxContainers int     `json:"max_containers,string"`
	Metric        string  `json:"metric"`
	Target        float64 `json:"target"`
	TargetText    string  `json:"target_text"`
}

type EventEditAutoscalerType struct {
	Event
	TypeData EventEditAutoscalerTypeData `json:"type_data"`
}

func (ev *EventEditAutoscalerType) String() string {
	d := ev.TypeData
	return fmt.Sprintf("Autoscaler edited about %s on container %s (target: %s)", d.Metric, d.ContainerType, d.TargetText)
}

type EventDeleteAutoscalerTypeData struct {
	ContainerType string `json:"container_type"`
	Metric        string `json:"metric"`
}

type EventDeleteAutoscalerType struct {
	Event
	TypeData EventDeleteAutoscalerTypeData `json:"type_data"`
}

func (ev *EventDeleteAutoscalerType) String() string {
	d := ev.TypeData
	return fmt.Sprintf("Alert deleted about %s on container %s", d.Metric, d.ContainerType)
}

type EventStartRegionMigrationTypeData struct {
	MigrationID string `json:"migration_id"`
	Destination string `json:"destination"`
	Source      string `json:"source"`
	DstAppName  string `json:"dst_app_name"`
}

type EventStartRegionMigrationType struct {
	Event
	TypeData EventStartRegionMigrationTypeData `json:"type_data"`
}

func (ev *EventStartRegionMigrationType) String() string {
	return fmt.Sprintf("Application region migration started from %s to %s/%s", ev.TypeData.Source, ev.TypeData.Destination, ev.TypeData.DstAppName)
}

// New log drain
type EventNewLogDrainTypeData struct {
	URL string `json:"url"`
}

type EventNewLogDrainType struct {
	Event
	TypeData EventNewLogDrainTypeData `json:"type_data"`
}

func (ev *EventNewLogDrainType) String() string {
	return fmt.Sprintf("Log drain added on %s app", ev.AppName)
}

// Delete log drain
type EventDeleteLogDrainTypeData struct {
	URL string `json:"url"`
}

type EventDeleteLogDrainType struct {
	Event
	TypeData EventDeleteLogDrainTypeData `json:"type_data"`
}

func (ev *EventDeleteLogDrainType) String() string {
	return fmt.Sprintf("Log drain deleted on %s app", ev.AppName)
}

// New addon log drain
type EventNewAddonLogDrainTypeData struct {
	URL       string `json:"url"`
	AddonUUID string `json:"addon_uuid"`
	AddonName string `json:"addon_name"`
}

type EventNewAddonLogDrainType struct {
	Event
	TypeData EventNewAddonLogDrainTypeData `json:"type_data"`
}

func (ev *EventNewAddonLogDrainType) String() string {
	return fmt.Sprintf("Log drain added for %s addon on %s app", ev.TypeData.AddonName, ev.AppName)
}

// Delete addon log drain
type EventDeleteAddonLogDrainTypeData struct {
	URL       string `json:"url"`
	AddonUUID string `json:"addon_uuid"`
	AddonName string `json:"addon_name"`
}

type EventDeleteAddonLogDrainType struct {
	Event
	TypeData EventDeleteAddonLogDrainTypeData `json:"type_data"`
}

func (ev *EventDeleteAddonLogDrainType) String() string {
	return fmt.Sprintf("Log drain deleted on %s addon for %s app", ev.TypeData.AddonName, ev.AppName)
}

// New notifier
type EventNewNotifierTypeData struct {
	NotifierName     string                 `json:"notifier_name"`
	Active           bool                   `json:"active"`
	SendAllEvents    bool                   `json:"send_all_events"`
	SelectedEvents   []string               `json:"selected_events"`
	NotifierType     string                 `json:"notifier_type"`
	NotifierTypeData map[string]interface{} `json:"notifier_type_data"`
	PlatformName     string                 `json:"platform_name"`
}

type EventNewNotifierType struct {
	Event
	TypeData EventNewNotifierTypeData `json:"type_data"`
}

var NotifierPlatformNames = map[string]string{
	"email":       "E-mail",
	"rocker_chat": "Rocket Chat",
	"slack":       "Slack",
	"webhook":     "Webhook",
}

func (ev *EventNewNotifierType) String() string {
	d := ev.TypeData
	platformName, ok := NotifierPlatformNames[d.PlatformName]
	if !ok {
		platformName = "unknown"
	}
	return fmt.Sprintf("Notifier '%s' created for the platform '%s' on %s app", d.NotifierName, platformName, ev.AppName)
}

// Edit notifier
type EventEditNotifierTypeData struct {
	NotifierName     string                 `json:"notifier_name"`
	Active           bool                   `json:"active"`
	SendAllEvents    bool                   `json:"send_all_events"`
	SelectedEvents   []string               `json:"selected_events"`
	NotifierType     string                 `json:"notifier_type"`
	NotifierTypeData map[string]interface{} `json:"notifier_type_data"`
	PlatformName     string                 `json:"platform_name"`
}

type EventEditNotifierType struct {
	Event
	TypeData EventEditNotifierTypeData `json:"type_data"`
}

func (ev *EventEditNotifierType) String() string {
	d := ev.TypeData
	return fmt.Sprintf("Notifier '%s' edited on %s app", d.NotifierName, ev.AppName)
}

// Delete notifier
type EventDeleteNotifierTypeData struct {
	NotifierName     string                 `json:"notifier_name"`
	Active           bool                   `json:"active"`
	SendAllEvents    bool                   `json:"send_all_events"`
	SelectedEvents   []string               `json:"selected_events"`
	NotifierType     string                 `json:"notifier_type"`
	NotifierTypeData map[string]interface{} `json:"notifier_type_data"`
	PlatformName     string                 `json:"platform_name"`
}

type EventDeleteNotifierType struct {
	Event
	TypeData EventDeleteNotifierTypeData `json:"type_data"`
}

func (ev *EventDeleteNotifierType) String() string {
	d := ev.TypeData
	return fmt.Sprintf("Notifier '%s' deleted on %s app", d.NotifierName, ev.AppName)
}

// Edit hds_contact
type EventEditHDSContactTypeData struct {
	Name           string `json:"name"`
	Email          string `json:"email"`
	PhoneNumber    string `json:"phone_number"`
	Company        string `json:"company"`
	AddressLine1   string `json:"address_line1"`
	AddressLine2   string `json:"address_line2"`
	AddressCity    string `json:"address_city"`
	AddressZip     string `json:"address_zip"`
	AddressCountry string `json:"address_country"`
	Notes          string `json:"notes"`
}

type EventEditHDSContactType struct {
	Event
	TypeData EventEditHDSContactTypeData `json:"type_data"`
}

func (ev *EventEditHDSContactType) String() string {
	d := ev.TypeData
	return fmt.Sprintf("Contact health Professional '%s' edited on %s app", d.Name, ev.AppName)
}

// Create data_access_consent
type EventCreateDataAccessConsentTypeData struct {
	EndAt      time.Time `json:"end_at"`
	Databases  bool      `json:"databases"`
	Containers bool      `json:"containers"`
}

type EventCreateDataAccessConsentType struct {
	Event
	TypeData EventCreateDataAccessConsentTypeData `json:"type_data"`
}

func (ev *EventCreateDataAccessConsentType) String() string {
	d := ev.TypeData
	res := "Additional access "
	if d.Containers {
		res += "to application runtime environment, "
	}
	if d.Databases {
		res += "to databases metadata and monitoring data, "
	}
	res += fmt.Sprintf("created on %s app", ev.AppName)
	return res
}

// Enable Tfa
type EventTfaEnabledTypeData struct {
	Provider string `json:"provider"`
}

type EventTfaEnabledType struct {
	Event
	TypeData EventTfaEnabledTypeData `json:"type_data"`
}

func (ev *EventTfaEnabledType) String() string {
	return fmt.Sprintf("Two factor authentication enabled by %s", ev.TypeData.Provider)
}

// Disable Tfa
type EventTfaDisabledTypeData struct {
}

type EventTfaDisabledType struct {
	Event
	TypeData EventTfaDisabledTypeData `json:"type_data"`
}

func (ev *EventTfaDisabledType) String() string {
	return "Two factor authentication disabled"
}

// Stack changed
type EventStackChangedTypeData struct {
	PreviousStackID   string `json:"previous_stack_id"`
	CurrentStackID    string `json:"current_stack_id"`
	PreviousStackName string `json:"previous_stack_name"`
	CurrentStackName  string `json:"current_stack_name"`
}

type EventStackChangedType struct {
	Event
	TypeData EventStackChangedTypeData `json:"type_data"`
}

func (ev *EventStackChangedType) String() string {
	d := ev.TypeData
	return fmt.Sprintf("Stack changed from '%s' to %s", d.PreviousStackName, d.CurrentStackName)
}

// Database maintenance planned
type EventPlanDatabaseMaintenanceTypeData struct {
	AddonName                string    `json:"addon_name"`
	MaintenanceID            string    `json:"maintenance_id"`
	MaintenanceWindowInHours int       `json:"maintenance_window_in_hours"`
	MaintenanceType          string    `json:"maintenance_type"`
	NextMaintenanceWindow    time.Time `json:"next_maintenance_window"`
}

type EventPlanDatabaseMaintenanceType struct {
	Event
	TypeData EventPlanDatabaseMaintenanceTypeData `json:"type_data"`
}

func (ev *EventPlanDatabaseMaintenanceType) String() string {
	return fmt.Sprintf("A maintenance (ID: %s) has been scheduled on the %s database.", ev.TypeData.MaintenanceID, ev.TypeData.AddonName)
}

func (ev *EventPlanDatabaseMaintenanceType) Who() string {
	return ev.Event.Who()
}

// Database maintenance started
type EventStartDatabaseMaintenanceTypeData struct {
	AddonName                string    `json:"addon_name"`
	MaintenanceID            string    `json:"maintenance_id"`
	MaintenanceWindowInHours int       `json:"maintenance_window_in_hours"`
	MaintenanceType          string    `json:"maintenance_type"`
	NextMaintenanceWindow    time.Time `json:"next_maintenance_window"`
}

type EventStartDatabaseMaintenanceType struct {
	Event
	TypeData EventStartDatabaseMaintenanceTypeData `json:"type_data"`
}

func (ev *EventStartDatabaseMaintenanceType) String() string {
	return fmt.Sprintf("A maintenance (ID: %s) has started on the %s database.", ev.TypeData.MaintenanceID, ev.TypeData.AddonName)
}

func (ev *EventStartDatabaseMaintenanceType) Who() string {
	return ev.Event.Who()
}

// Database maintenance completed
type EventCompleteDatabaseMaintenanceTypeData struct {
	AddonName                string    `json:"addon_name"`
	MaintenanceID            string    `json:"maintenance_id"`
	MaintenanceWindowInHours int       `json:"maintenance_window_in_hours"`
	MaintenanceType          string    `json:"maintenance_type"`
	NextMaintenanceWindow    time.Time `json:"next_maintenance_window"`
}

type EventCompleteDatabaseMaintenanceType struct {
	Event
	TypeData EventCompleteDatabaseMaintenanceTypeData `json:"type_data"`
}

func (ev *EventCompleteDatabaseMaintenanceType) String() string {
	return fmt.Sprintf("A maintenance (ID: %s) has been completed on the %s database.", ev.TypeData.MaintenanceID, ev.TypeData.AddonName)
}

func (ev *EventCompleteDatabaseMaintenanceType) Who() string {
	return ev.Event.Who()
}

// Project deleted
type EventDeleteProjectTypeData struct {
}

type EventDeleteProjectType struct {
	Event
	TypeData EventDeleteProjectTypeData `json:"type_data"`
}

func (ev *EventDeleteProjectType) String() string {
	return fmt.Sprintf("The project '%s' has been deleted", ev.ProjectName)
}

// New project created
type EventNewProjectTypeData struct {
	Default bool `json:"default"`
}

type EventNewProjectType struct {
	Event
	TypeData EventNewProjectTypeData `json:"type_data"`
}

func (ev *EventNewProjectType) String() string {
	return fmt.Sprintf("The project '%s' has been created", ev.ProjectName)
}

// Project edited
type EditProjectValue struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	OldValue string `json:"old_value"`
}

type EditProjectValues []EditProjectValue

type EventEditProjectTypeData struct {
	UpdatedValues EditProjectValues `json:"updated_values"`
}

type EventEditProjectType struct {
	Event
	TypeData EventEditProjectTypeData `json:"type_data"`
}

func (ev *EventEditProjectType) String() string {
	changes := []string{}

	for _, v := range ev.TypeData.UpdatedValues {
		changes = append(changes, fmt.Sprintf("%s modified from '%v' to '%v'", v.Name, v.OldValue, v.Value))
	}

	return fmt.Sprintf("project settings have been updated: %s", strings.Join(changes, ", "))
}
