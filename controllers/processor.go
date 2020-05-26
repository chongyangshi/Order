package controllers

// The processor is responsible for periodically inspecting managed resources on the
// buffer, and check whether all their depending pod controllers are up-to-date in
// terms of restarts. If any isn't and they can be restarted based on the cooldown
// configured, then the procesor will apply an annotation to ask Kubernetes to restart
// the said pod controller.
