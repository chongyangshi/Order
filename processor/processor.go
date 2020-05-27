package processor

// @TODO: Control loop
// - Pop from buffer
// - Check whether resource is managed
// - If resource is not managed, discard
// - If resource is managed, continue
// - Push back into buffer if minimum cooldown not yet met
// - Look up currently matching pod controllers from cachers
// - For each matching pod controller, if namespace does not qualify, skip
// - If namespace qualifies, check whether managed resource hash annotations match
// - If matches, skip as pod controller is already up to date
// - If does not match, check if minimum cooldown from last restart qualifies
// - If minimum cooldown does not qualifies, push back into buffer and skip
// - If minimum cooldown qualifies, perform rolling restart and continue.
