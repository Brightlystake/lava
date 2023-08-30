package keeper_test

import (
	"fmt"
	"math"
	"sort"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lavanet/lava/testutil/common"
	testkeeper "github.com/lavanet/lava/testutil/keeper"
	"github.com/lavanet/lava/utils/slices"
	epochstoragetypes "github.com/lavanet/lava/x/epochstorage/types"
	pairingscores "github.com/lavanet/lava/x/pairing/keeper/scores"
	planstypes "github.com/lavanet/lava/x/plans/types"
	spectypes "github.com/lavanet/lava/x/spec/types"
	"github.com/stretchr/testify/require"
)

func TestPairingUniqueness(t *testing.T) {
	ts := newTester(t)
	ts.SetupAccounts(2, 0, 0) // 2 sub, 0 adm, 0 dev

	var balance int64 = 10000
	stake := balance / 10

	_, sub1Addr := ts.Account("sub1")
	_, sub2Addr := ts.Account("sub2")

	_, err := ts.TxSubscriptionBuy(sub1Addr, sub1Addr, ts.plan.Index, 1)
	require.Nil(t, err)
	_, err = ts.TxSubscriptionBuy(sub2Addr, sub2Addr, ts.plan.Index, 1)
	require.Nil(t, err)

	for i := 1; i <= 1000; i++ {
		_, addr := ts.AddAccount(common.PROVIDER, i, balance)
		err := ts.StakeProvider(addr, ts.spec, stake)
		require.Nil(t, err)
	}

	ts.AdvanceEpoch()

	// test that 2 different clients get different pairings
	pairing1, err := ts.QueryPairingGetPairing(ts.spec.Index, sub1Addr)
	require.Nil(t, err)
	pairing2, err := ts.QueryPairingGetPairing(ts.spec.Index, sub2Addr)
	require.Nil(t, err)

	filter := func(p epochstoragetypes.StakeEntry) string { return p.Address }

	providerAddrs1 := slices.Filter(pairing1.Providers, filter)
	providerAddrs2 := slices.Filter(pairing2.Providers, filter)

	require.Equal(t, len(pairing1.Providers), len(pairing2.Providers))
	require.False(t, slices.UnorderedEqual(providerAddrs1, providerAddrs2))

	ts.AdvanceEpoch()

	// test that in different epoch we get different pairings for consumer1
	pairing11, err := ts.QueryPairingGetPairing(ts.spec.Index, sub1Addr)
	require.Nil(t, err)

	providerAddrs11 := slices.Filter(pairing11.Providers, filter)

	require.Equal(t, len(pairing1.Providers), len(pairing11.Providers))
	require.False(t, slices.UnorderedEqual(providerAddrs1, providerAddrs11))

	// test that get pairing gives the same results for the whole epoch
	epochBlocks := ts.EpochBlocks()
	for i := uint64(0); i < epochBlocks-1; i++ {
		ts.AdvanceBlock()

		pairing111, err := ts.QueryPairingGetPairing(ts.spec.Index, sub1Addr)
		require.Nil(t, err)

		for i := range pairing11.Providers {
			providerAddr := pairing11.Providers[i].Address
			require.Equal(t, providerAddr, pairing111.Providers[i].Address)
			require.Nil(t, err)
			verify, err := ts.QueryPairingVerifyPairing(ts.spec.Index, sub1Addr, providerAddr, ts.BlockHeight())
			require.Nil(t, err)
			require.True(t, verify.Valid)
		}
	}
}

func TestValidatePairingDeterminism(t *testing.T) {
	ts := newTester(t)
	ts.SetupAccounts(1, 0, 0) // 1 sub, 0 adm, 0 dev

	var balance int64 = 10000
	stake := balance / 10

	_, sub1Addr := ts.Account("sub1")

	_, err := ts.TxSubscriptionBuy(sub1Addr, sub1Addr, ts.plan.Index, 1)
	require.Nil(t, err)

	for i := 1; i <= 10; i++ {
		_, addr := ts.AddAccount(common.PROVIDER, i, balance)
		err := ts.StakeProvider(addr, ts.spec, stake)
		require.Nil(t, err)
	}

	ts.AdvanceEpoch()

	// test that 2 different clients get different pairings
	pairing, err := ts.QueryPairingGetPairing(ts.spec.Index, sub1Addr)
	require.Nil(t, err)

	block := ts.BlockHeight()
	testAllProviders := func() {
		for _, provider := range pairing.Providers {
			providerAddr := provider.Address
			verify, err := ts.QueryPairingVerifyPairing(ts.spec.Index, sub1Addr, providerAddr, block)
			require.Nil(t, err)
			require.True(t, verify.Valid)
		}
	}

	count := ts.BlocksToSave() - ts.BlockHeight()
	for i := 0; i < int(count); i++ {
		ts.AdvanceBlock()
		testAllProviders()
	}
}

func TestGetPairing(t *testing.T) {
	ts := newTester(t)

	// do not use ts.setupForPayments(1, 1, 0), because it kicks off with AdvanceEpoch()
	// (for the benefit of users) but the "zeroEpoch" test below expects to start at the
	// same epoch of staking the providers.
	ts.addClient(1)
	err := ts.addProvider(1)
	require.Nil(t, err)

	// BLOCK_TIME = 30sec (testutil/keeper/keepers_init.go)
	constBlockTime := testkeeper.BLOCK_TIME
	epochBlocks := ts.EpochBlocks()

	// test: different epoch, valid tells if the payment request should work
	tests := []struct {
		name                string
		validPairingExists  bool
		isEpochTimesChanged bool
	}{
		{"zeroEpoch", false, false},
		{"firstEpoch", true, false},
		{"commonEpoch", true, false},
		{"epochTimesChanged", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Advance an epoch according to the test
			switch tt.name {
			case "zeroEpoch":
				// do nothing
			case "firstEpoch":
				ts.AdvanceEpoch()
			case "commonEpoch":
				ts.AdvanceEpochs(5)
			case "epochTimesChanged":
				ts.AdvanceEpochs(5)
				ts.AdvanceBlocks(epochBlocks/2, constBlockTime/2)
				ts.AdvanceBlocks(epochBlocks / 2)
			}

			_, clientAddr := ts.GetAccount(common.CONSUMER, 0)
			_, providerAddr := ts.GetAccount(common.PROVIDER, 0)

			// get pairing for client (for epoch zero expect to fail)
			pairing, err := ts.QueryPairingGetPairing(ts.spec.Index, clientAddr)
			if !tt.validPairingExists {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)

				// verify the expected provider
				require.Equal(t, providerAddr, pairing.Providers[0].Address)

				// verify the current epoch
				epochThis := ts.EpochStart()
				require.Equal(t, epochThis, pairing.CurrentEpoch)

				// verify the SpecLastUpdatedBlock
				require.Equal(t, ts.spec.BlockLastUpdated, pairing.SpecLastUpdatedBlock)

				// if prevEpoch == 0 -> averageBlockTime = 0
				// else calculate the time (like the actual get-pairing function)
				epochPrev, err := ts.Keepers.Epochstorage.GetPreviousEpochStartForBlock(ts.Ctx, epochThis)
				require.Nil(t, err)

				var averageBlockTime uint64
				if epochPrev != 0 {
					// calculate average block time base on total time from first block of
					// previous epoch until first block of this epoch and block dfference.
					blockCore1 := ts.Keepers.BlockStore.LoadBlock(int64(epochPrev))
					blockCore2 := ts.Keepers.BlockStore.LoadBlock(int64(epochThis))
					delta := blockCore2.Time.Sub(blockCore1.Time).Seconds()
					averageBlockTime = uint64(delta) / (epochThis - epochPrev)
				}

				overlapBlocks := ts.Keepers.Pairing.EpochBlocksOverlap(ts.Ctx)
				nextEpochStart, err := ts.Keepers.Epochstorage.GetNextEpoch(ts.Ctx, epochThis)
				require.Nil(t, err)

				// calculate the block in which the next pairing will happen (+overlap)
				nextPairingBlock := nextEpochStart + overlapBlocks
				// calculate number of blocks from the current block to the next epoch
				blocksUntilNewEpoch := nextPairingBlock - ts.BlockHeight()
				// calculate time left for the next pairing (blocks left* avg block time)
				timeLeftToNextPairing := blocksUntilNewEpoch * averageBlockTime

				if !tt.isEpochTimesChanged {
					require.Equal(t, timeLeftToNextPairing, pairing.TimeLeftToNextPairing)
				} else {
					// averageBlockTime in get-pairing query -> minimal average across sampled epoch
					// averageBlockTime in this test -> normal average across epoch
					// we've used a smaller blocktime some of the time -> averageBlockTime from
					// get-pairing is smaller than the averageBlockTime calculated in this test
					require.Less(t, pairing.TimeLeftToNextPairing, timeLeftToNextPairing)
				}

				// verify nextPairingBlock
				require.Equal(t, nextPairingBlock, pairing.BlockOfNextPairing)
			}
		})
	}
}

func TestPairingStatic(t *testing.T) {
	ts := newTester(t)
	ts.SetupAccounts(1, 0, 0) // 1 sub, 0 adm, 0 dev

	_, sub1Addr := ts.Account("sub1")

	ts.spec.ProvidersTypes = spectypes.Spec_static
	// will overwrite the default "mock" spec
	// (no TxProposalAddSpecs because the mock spec does not pass validaton)
	ts.AddSpec("mock", ts.spec)

	ts.AdvanceEpoch()

	_, err := ts.TxSubscriptionBuy(sub1Addr, sub1Addr, ts.plan.Index, 1)
	require.Nil(t, err)

	for i := 0; i < int(ts.plan.PlanPolicy.MaxProvidersToPair)*2; i++ {
		_, addr := ts.AddAccount(common.PROVIDER, i, testBalance)
		err := ts.StakeProvider(addr, ts.spec, testStake+int64(i))
		require.Nil(t, err)
	}

	// we expect to get all the providers in static spec

	ts.AdvanceEpoch()

	pairing, err := ts.QueryPairingGetPairing(ts.spec.Index, sub1Addr)
	require.Nil(t, err)

	for i, provider := range pairing.Providers {
		require.Equal(t, provider.Stake.Amount.Int64(), testStake+int64(i))
	}
}

func TestAddonPairing(t *testing.T) {
	ts := newTester(t)
	ts.setupForPayments(1, 0, 0) // 1 provider, 0 client, default providers-to-pair

	mandatory := spectypes.CollectionData{
		ApiInterface: "mandatory",
		InternalPath: "",
		Type:         "",
		AddOn:        "",
	}
	mandatoryAddon := spectypes.CollectionData{
		ApiInterface: "mandatory",
		InternalPath: "",
		Type:         "",
		AddOn:        "addon",
	}
	optional := spectypes.CollectionData{
		ApiInterface: "optional",
		InternalPath: "",
		Type:         "",
		AddOn:        "optional",
	}
	ts.spec.ApiCollections = []*spectypes.ApiCollection{
		{
			Enabled:        true,
			CollectionData: mandatory,
		},
		{
			Enabled:        true,
			CollectionData: optional,
		},
		{
			Enabled:        true,
			CollectionData: mandatoryAddon,
		},
	}

	// will overwrite the default "mock" spec
	ts.AddSpec("mock", ts.spec)
	specId := ts.spec.Index

	mandatoryChainPolicy := &planstypes.ChainPolicy{
		ChainId:      specId,
		Requirements: []planstypes.ChainRequirement{{Collection: mandatory}},
	}
	mandatoryAddonChainPolicy := &planstypes.ChainPolicy{
		ChainId:      specId,
		Requirements: []planstypes.ChainRequirement{{Collection: mandatoryAddon}},
	}
	optionalAddonChainPolicy := &planstypes.ChainPolicy{
		ChainId:      specId,
		Requirements: []planstypes.ChainRequirement{{Collection: optional}},
	}
	optionalAndMandatoryAddonChainPolicy := &planstypes.ChainPolicy{
		ChainId:      specId,
		Requirements: []planstypes.ChainRequirement{{Collection: mandatoryAddon}, {Collection: optional}},
	}

	templates := []struct {
		name                      string
		planChainPolicy           *planstypes.ChainPolicy
		subscChainPolicy          *planstypes.ChainPolicy
		projChainPolicy           *planstypes.ChainPolicy
		expectedProviders         int
		expectedStrictestPolicies []string
	}{
		{
			name:              "empty",
			expectedProviders: 12,
		},
		{
			name:              "mandatory in plan",
			planChainPolicy:   mandatoryChainPolicy,
			expectedProviders: 11,
		},
		{
			name:              "mandatory in subsc",
			subscChainPolicy:  mandatoryChainPolicy,
			projChainPolicy:   nil,
			expectedProviders: 11,
		},
		{
			name:              "mandatory in proj",
			projChainPolicy:   mandatoryChainPolicy,
			expectedProviders: 11,
		},
		{
			name:                      "addon in plan",
			planChainPolicy:           mandatoryAddonChainPolicy,
			subscChainPolicy:          nil,
			projChainPolicy:           nil,
			expectedProviders:         6,
			expectedStrictestPolicies: []string{"addon"},
		},
		{
			name:                      "addon in subsc",
			subscChainPolicy:          mandatoryAddonChainPolicy,
			expectedProviders:         6,
			expectedStrictestPolicies: []string{"addon"},
		},
		{
			name:                      "addon in proj",
			projChainPolicy:           mandatoryAddonChainPolicy,
			expectedProviders:         6,
			expectedStrictestPolicies: []string{"addon"},
		},
		{
			name:                      "optional in plan",
			planChainPolicy:           optionalAddonChainPolicy,
			expectedProviders:         7,
			expectedStrictestPolicies: []string{"optional"},
		},
		{
			name:                      "optional in subsc",
			subscChainPolicy:          optionalAddonChainPolicy,
			expectedProviders:         7,
			expectedStrictestPolicies: []string{"optional"},
		},
		{
			name:                      "optional in proj",
			projChainPolicy:           optionalAddonChainPolicy,
			expectedProviders:         7,
			expectedStrictestPolicies: []string{"optional"},
		},
		{
			name:                      "optional and addon in plan",
			planChainPolicy:           optionalAndMandatoryAddonChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
		{
			name:                      "optional and addon in subsc",
			subscChainPolicy:          optionalAndMandatoryAddonChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
		{
			name:                      "optional and addon in proj",
			projChainPolicy:           optionalAndMandatoryAddonChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
		{
			name:                      "optional and addon in plan, addon in subsc",
			planChainPolicy:           optionalAndMandatoryAddonChainPolicy,
			subscChainPolicy:          mandatoryAddonChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
		{
			name:                      "optional and addon in subsc, addon in plan",
			planChainPolicy:           mandatoryAddonChainPolicy,
			subscChainPolicy:          optionalAndMandatoryAddonChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
		{
			name:                      "optional and addon in subsc, addon in proj",
			subscChainPolicy:          optionalAndMandatoryAddonChainPolicy,
			projChainPolicy:           mandatoryAddonChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
		{
			name:                      "optional in subsc, addon in proj",
			subscChainPolicy:          optionalAndMandatoryAddonChainPolicy,
			projChainPolicy:           mandatoryAddonChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
	}

	mandatorySupportingEndpoints := []epochstoragetypes.Endpoint{{
		IPPORT:        "123",
		Geolocation:   1,
		Addons:        []string{mandatory.AddOn},
		ApiInterfaces: []string{mandatory.ApiInterface},
	}}
	mandatoryAddonSupportingEndpoints := []epochstoragetypes.Endpoint{{
		IPPORT:        "456",
		Geolocation:   1,
		Addons:        []string{mandatoryAddon.AddOn},
		ApiInterfaces: []string{mandatoryAddon.ApiInterface},
	}}
	mandatoryAndMandatoryAddonSupportingEndpoints := slices.Concat(
		mandatorySupportingEndpoints, mandatoryAddonSupportingEndpoints)

	optionalSupportingEndpoints := []epochstoragetypes.Endpoint{{
		IPPORT:        "789",
		Geolocation:   1,
		Addons:        []string{optional.AddOn},
		ApiInterfaces: []string{optional.ApiInterface},
	}}
	optionalAndMandatorySupportingEndpoints := slices.Concat(
		mandatorySupportingEndpoints, optionalSupportingEndpoints)
	optionalAndMandatoryAddonSupportingEndpoints := slices.Concat(
		mandatoryAddonSupportingEndpoints, optionalSupportingEndpoints)

	allSupportingEndpoints := slices.Concat(
		mandatorySupportingEndpoints, optionalAndMandatoryAddonSupportingEndpoints)

	mandatoryAndOptionalSingleEndpoint := []epochstoragetypes.Endpoint{{
		IPPORT:        "444",
		Geolocation:   1,
		Addons:        []string{},
		ApiInterfaces: []string{mandatoryAddon.ApiInterface, optional.ApiInterface},
	}}

	err := ts.addProviderEndpoints(2, mandatorySupportingEndpoints)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(2, mandatoryAndMandatoryAddonSupportingEndpoints)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(2, optionalAndMandatorySupportingEndpoints)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(1, mandatoryAndOptionalSingleEndpoint)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(2, optionalAndMandatoryAddonSupportingEndpoints)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(2, allSupportingEndpoints)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(2, optionalSupportingEndpoints) // this errors out
	require.Error(t, err)

	stakeStorage, found := ts.Keepers.Epochstorage.GetStakeStorageCurrent(ts.Ctx, ts.spec.Index)
	require.True(t, found)
	require.Len(t, stakeStorage.StakeEntries, 12)

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			defaultPolicy := func() planstypes.Policy {
				return planstypes.Policy{
					ChainPolicies:      []planstypes.ChainPolicy{},
					GeolocationProfile: math.MaxUint64,
					MaxProvidersToPair: 100,
					TotalCuLimit:       math.MaxUint64,
					EpochCuLimit:       math.MaxUint64,
				}
			}

			plan := ts.plan // original mock template
			plan.PlanPolicy = defaultPolicy()

			if tt.planChainPolicy != nil {
				plan.PlanPolicy.ChainPolicies = []planstypes.ChainPolicy{*tt.planChainPolicy}
			}

			err := ts.TxProposalAddPlans(plan)
			require.Nil(t, err)

			_, sub1Addr := ts.AddAccount("sub", 0, 10000)

			_, err = ts.TxSubscriptionBuy(sub1Addr, sub1Addr, plan.Index, 1)
			require.Nil(t, err)

			// get the admin project and set its policies
			subProjects, err := ts.QuerySubscriptionListProjects(sub1Addr)
			require.Nil(t, err)
			require.Equal(t, 1, len(subProjects.Projects))

			projectID := subProjects.Projects[0]

			if tt.projChainPolicy != nil {
				projPolicy := defaultPolicy()
				projPolicy.ChainPolicies = []planstypes.ChainPolicy{*tt.projChainPolicy}
				_, err = ts.TxProjectSetPolicy(projectID, sub1Addr, projPolicy)
				require.Nil(t, err)
			}

			// apply policy change
			ts.AdvanceEpoch()

			if tt.subscChainPolicy != nil {
				subscPolicy := defaultPolicy()
				subscPolicy.ChainPolicies = []planstypes.ChainPolicy{*tt.subscChainPolicy}
				_, err = ts.TxProjectSetSubscriptionPolicy(projectID, sub1Addr, subscPolicy)
				require.Nil(t, err)
			}

			// apply policy change
			ts.AdvanceEpochs(2)

			project, err := ts.GetProjectForBlock(projectID, ts.BlockHeight())
			require.NoError(t, err)

			strictestPolicy, _, err := ts.Keepers.Pairing.GetProjectStrictestPolicy(ts.Ctx, project, specId)
			require.NoError(t, err)
			if len(tt.expectedStrictestPolicies) > 0 {
				require.NotEqual(t, 0, len(strictestPolicy.ChainPolicies))
				require.NotEqual(t, 0, len(strictestPolicy.ChainPolicies[0].Requirements))
				addons := map[string]struct{}{}
				for _, requirement := range strictestPolicy.ChainPolicies[0].Requirements {
					collection := requirement.Collection
					if collection.AddOn != "" {
						addons[collection.AddOn] = struct{}{}
					}
				}
				for _, expected := range tt.expectedStrictestPolicies {
					_, ok := addons[expected]
					require.True(t, ok, "did not find addon in strictest policy %s, policy: %#v", expected, strictestPolicy)
				}
			}

			pairing, err := ts.QueryPairingGetPairing(ts.spec.Index, sub1Addr)
			if tt.expectedProviders > 0 {
				require.Nil(t, err)
				require.Equal(t, tt.expectedProviders, len(pairing.Providers), "received providers %#v", pairing)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func countSelectedAddresses(selected []string, expected []string) int {
	count := 0
	countPossibilities := map[string]struct{}{}
	for _, possibility := range expected {
		countPossibilities[possibility] = struct{}{}
	}
	for _, selectedProvider := range selected {
		_, ok := countPossibilities[selectedProvider]
		if ok {
			count++
		}
	}
	return count
}

func TestSelectedProvidersPairing(t *testing.T) {
	ts := newTester(t)

	err := ts.addProvider(200)
	require.Nil(t, err)

	policy := &planstypes.Policy{
		GeolocationProfile: math.MaxUint64,
		MaxProvidersToPair: 3,
	}

	allowed := planstypes.SELECTED_PROVIDERS_MODE_ALLOWED
	exclusive := planstypes.SELECTED_PROVIDERS_MODE_EXCLUSIVE
	disabled := planstypes.SELECTED_PROVIDERS_MODE_DISABLED
	mixed := planstypes.SELECTED_PROVIDERS_MODE_MIXED

	maxProvidersToPair, err := ts.Keepers.Pairing.CalculateEffectiveProvidersToPairFromPolicies(
		[]*planstypes.Policy{&ts.plan.PlanPolicy, policy},
	)
	require.Nil(t, err)

	err = ts.addProvider(200)
	require.Nil(t, err)

	_, p1 := ts.GetAccount(common.PROVIDER, 0)
	_, p2 := ts.GetAccount(common.PROVIDER, 1)
	_, p3 := ts.GetAccount(common.PROVIDER, 2)
	_, p4 := ts.GetAccount(common.PROVIDER, 3)
	_, p5 := ts.GetAccount(common.PROVIDER, 4)

	providerSets := []struct {
		planProviders []string
		subProviders  []string
		projProviders []string
	}{
		{[]string{}, []string{}, []string{}},                                 // set #0
		{[]string{p1, p2, p3}, []string{}, []string{}},                       // set #1
		{[]string{p1, p2}, []string{}, []string{}},                           // set #2
		{[]string{p3, p4}, []string{}, []string{}},                           // set #3
		{[]string{p1, p2, p3}, []string{p1, p2}, []string{}},                 // set #4
		{[]string{p1, p2, p3}, []string{}, []string{p1, p3}},                 // set #5
		{[]string{}, []string{p1, p2, p3}, []string{p1, p2}},                 // set #6
		{[]string{p1}, []string{p1, p2, p3}, []string{p1, p2}},               // set #7
		{[]string{p1, p2, p3, p4, p5}, []string{p1, p2, p3, p4}, []string{}}, // set #8
	}

	expectedSelectedProviders := [][]string{
		{p1, p2, p3},     // expected providers for intersection of set #1
		{p1, p2},         // expected providers for intersection of set #2
		{p3, p4},         // expected providers for intersection of set #3
		{p1, p2},         // expected providers for intersection of set #4
		{p1, p3},         // expected providers for intersection of set #5
		{p1, p2},         // expected providers for intersection of set #6
		{p1},             // expected providers for intersection of set #7
		{p1, p2, p3, p4}, // expected providers for intersection of set #8
	}

	// TODO: add mixed mode test cases (once implemented)
	templates := []struct {
		name              string
		planMode          planstypes.SELECTED_PROVIDERS_MODE
		subMode           planstypes.SELECTED_PROVIDERS_MODE
		projMode          planstypes.SELECTED_PROVIDERS_MODE
		providersSet      int
		expectedProviders int
	}{
		// normal pairing cases
		{"ALLOWED mode normal pairing", allowed, allowed, allowed, 0, 0},
		{"DISABLED mode normal pairing", disabled, allowed, allowed, 0, 0},

		// basic pairing checks cases
		{"EXCLUSIVE mode selected MaxProvidersToPair providers", exclusive, allowed, allowed, 1, 0},
		{"EXCLUSIVE mode selected less than MaxProvidersToPair providers", exclusive, allowed, allowed, 2, 1},
		{"EXCLUSIVE mode selected less than MaxProvidersToPair different providers", exclusive, allowed, allowed, 3, 2},

		// intersection checks cases
		{"EXCLUSIVE mode intersection between plan/sub policies", exclusive, exclusive, exclusive, 4, 3},
		{"EXCLUSIVE mode intersection between plan/proj policies", exclusive, exclusive, exclusive, 5, 4},
		{"EXCLUSIVE mode intersection between sub/proj policies", exclusive, exclusive, exclusive, 6, 5},
		{"EXCLUSIVE mode intersection between all policies", exclusive, exclusive, exclusive, 7, 6},

		// selected providers more than MaxProvidersToPair
		{"EXCLUSIVE mode selected more than MaxProvidersToPair providers", exclusive, exclusive, exclusive, 8, 7},

		// provider unstake checks cases
		{"EXCLUSIVE mode provider unstakes after first pairing", exclusive, exclusive, exclusive, 1, 0},
		{"EXCLUSIVE mode non-staked provider stakes after first pairing", exclusive, exclusive, exclusive, 1, 0},

		{"MIXED mode normal pairing", mixed, mixed, mixed, 0, 0},
		{"MIXED mode pairing", mixed, mixed, mixed, 1, 1},
		{"MIXED mode intersection between plan/sub policies", mixed, mixed, mixed, 4, 3},
		{"MIXED mode intersection between plan/proj policies", mixed, mixed, mixed, 5, 4},
		{"MIXED mode intersection between sub/proj policies", mixed, mixed, mixed, 6, 5},
		{"MIXED mode intersection between all policies", mixed, mixed, mixed, 7, 6},
	}

	var expectedProvidersAfterUnstake []string

	for i, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			_, sub1Addr := ts.AddAccount("sub", i, 10000)

			// create plan, propose it and buy subscription
			plan := ts.Plan("mock")
			providersSet := providerSets[tt.providersSet]

			plan.PlanPolicy.SelectedProvidersMode = tt.planMode
			plan.PlanPolicy.SelectedProviders = providersSet.planProviders

			err := ts.TxProposalAddPlans(plan)
			require.Nil(t, err)

			_, err = ts.TxSubscriptionBuy(sub1Addr, sub1Addr, plan.Index, 1)
			require.Nil(t, err)

			// get the admin project and set its policies
			res, err := ts.QuerySubscriptionListProjects(sub1Addr)
			require.Nil(t, err)
			require.Equal(t, 1, len(res.Projects))

			project, err := ts.GetProjectForBlock(res.Projects[0], ts.BlockHeight())
			require.Nil(t, err)

			policy.SelectedProvidersMode = tt.projMode
			policy.SelectedProviders = providersSet.projProviders

			_, err = ts.TxProjectSetPolicy(project.Index, sub1Addr, *policy)
			require.Nil(t, err)

			// skip epoch for the policy change to take effect
			ts.AdvanceEpoch()

			policy.SelectedProvidersMode = tt.subMode
			policy.SelectedProviders = providersSet.subProviders

			_, err = ts.TxProjectSetSubscriptionPolicy(project.Index, sub1Addr, *policy)
			require.Nil(t, err)

			// skip epoch for the policy change to take effect
			ts.AdvanceEpoch()
			// and another epoch to get pairing of two consecutive epochs
			ts.AdvanceEpoch()

			pairing, err := ts.QueryPairingGetPairing(ts.spec.Index, sub1Addr)
			require.Nil(t, err)

			providerAddresses1 := []string{}
			for _, provider := range pairing.Providers {
				providerAddresses1 = append(providerAddresses1, provider.Address)
			}

			if tt.name == "EXCLUSIVE mode provider unstakes after first pairing" {
				// unstake p1 and remove from expected providers
				_, err = ts.TxPairingUnstakeProvider(p1, ts.spec.Index)
				require.Nil(t, err)
				expectedProvidersAfterUnstake = expectedSelectedProviders[tt.expectedProviders][1:]
			} else if tt.name == "EXCLUSIVE mode non-staked provider stakes after first pairing" {
				err := ts.StakeProvider(p1, ts.spec, 10000000)
				require.Nil(t, err)
			}

			ts.AdvanceEpoch()

			pairing, err = ts.QueryPairingGetPairing(ts.spec.Index, sub1Addr)
			require.Nil(t, err)

			providerAddresses2 := []string{}
			for _, provider := range pairing.Providers {
				providerAddresses2 = append(providerAddresses2, provider.Address)
			}

			// check pairings
			switch tt.name {
			case "MIXED mode pairing",
				"MIXED mode intersection between plan/sub policies",
				"MIXED mode intersection between plan/proj policies",
				"MIXED mode intersection between sub/proj policies",
				"MIXED mode intersection between all policies":
				count := countSelectedAddresses(providerAddresses1, expectedSelectedProviders[tt.expectedProviders])
				require.GreaterOrEqual(t, count, len(providerAddresses1)/2)
			case "ALLOWED mode normal pairing", "DISABLED mode normal pairing":
				require.False(t, slices.UnorderedEqual(providerAddresses1, providerAddresses2))
				require.Equal(t, maxProvidersToPair, uint64(len(providerAddresses1)))
				require.Equal(t, maxProvidersToPair, uint64(len(providerAddresses2)))

			case "EXCLUSIVE mode selected MaxProvidersToPair providers":
				require.True(t, slices.UnorderedEqual(providerAddresses1, providerAddresses2))
				require.Equal(t, maxProvidersToPair, uint64(len(providerAddresses2)))
				require.True(t, slices.UnorderedEqual(expectedSelectedProviders[tt.expectedProviders], providerAddresses1))

			case "EXCLUSIVE mode selected less than MaxProvidersToPair providers",
				"EXCLUSIVE mode selected less than MaxProvidersToPair different providers",
				"EXCLUSIVE mode intersection between plan/sub policies",
				"EXCLUSIVE mode intersection between plan/proj policies",
				"EXCLUSIVE mode intersection between sub/proj policies",
				"EXCLUSIVE mode intersection between all policies":
				require.True(t, slices.UnorderedEqual(providerAddresses1, providerAddresses2))
				require.Less(t, uint64(len(providerAddresses1)), maxProvidersToPair)
				require.True(t, slices.UnorderedEqual(expectedSelectedProviders[tt.expectedProviders], providerAddresses1))
			case "EXCLUSIVE mode selected more than MaxProvidersToPair providers":
				require.True(t, slices.IsSubset(providerAddresses1, expectedSelectedProviders[tt.expectedProviders]))
				require.True(t, slices.IsSubset(providerAddresses2, expectedSelectedProviders[tt.expectedProviders]))
				require.Equal(t, maxProvidersToPair, uint64(len(providerAddresses1)))
				require.Equal(t, maxProvidersToPair, uint64(len(providerAddresses2)))

			case "EXCLUSIVE mode provider unstakes after first pairing":
				require.False(t, slices.UnorderedEqual(providerAddresses1, providerAddresses2))
				require.True(t, slices.UnorderedEqual(expectedSelectedProviders[tt.expectedProviders], providerAddresses1))
				require.True(t, slices.UnorderedEqual(expectedProvidersAfterUnstake, providerAddresses2))

			case "EXCLUSIVE mode non-staked provider stakes after first pairing":
				require.False(t, slices.UnorderedEqual(providerAddresses1, providerAddresses2))
				require.True(t, slices.UnorderedEqual(expectedSelectedProviders[tt.expectedProviders], providerAddresses2))
				require.True(t, slices.UnorderedEqual(expectedProvidersAfterUnstake, providerAddresses1))
			}
		})
	}
}

func (ts *tester) verifyPairingDistribution(desc, client string, providersToPair int, weight func(epochstoragetypes.StakeEntry) int64) {
	const iterations = 10000
	const epsilon = 0.15

	res, err := ts.QueryPairingProviders(ts.spec.Index, false)
	require.NoError(ts.T, err, desc)
	allProviders := res.StakeEntry

	// mapping from provider (address) to its index
	mapProviders := make(map[string]int)
	for i, provider := range allProviders {
		mapProviders[provider.Address] = i
	}

	// calculate the total expected weight
	var totalWeight int64
	for _, provider := range allProviders {
		totalWeight += weight(provider)
	}

	// calculate the occurrence histogram
	histogram := make(map[int]int64)
	for i := 0; i < iterations; i++ {
		res, err := ts.QueryPairingGetPairing(ts.spec.Index, client)
		require.NoError(ts.T, err, desc)

		for _, provider := range res.Providers {
			count := histogram[mapProviders[provider.Address]]
			histogram[mapProviders[provider.Address]] = count + 1
		}

		// advance epoch to to switch pairing
		ts.AdvanceEpoch()
	}

	// Check that the count for each provider aligns with their stake's probability
	for i, actual := range histogram {
		// calculate expected occurrences based on weight function
		weight := weight(allProviders[i])
		expect := (int64(providersToPair) * weight * iterations) / totalWeight
		require.InEpsilon(ts.T, expect, actual, epsilon,
			desc+fmt.Sprintf(" expect/actual %d/%d", expect, actual))
	}
}

// Test that the pairing process picks identical providers uniformly
func TestPairingUniformDistribution(t *testing.T) {
	providersCount := 10
	providersToPair := 3

	ts := newTester(t)
	ts.setupForPayments(providersCount, 1, providersToPair)
	_, clientAddr := ts.GetAccount(common.CONSUMER, 0)

	// extend the subscription to accommodate many (pairing) epochs
	_, err := ts.TxSubscriptionBuy(clientAddr, clientAddr, ts.plan.Index, 5)
	require.NoError(t, err)

	weightFunc := func(p epochstoragetypes.StakeEntry) int64 { return p.Stake.Amount.Int64() }
	ts.verifyPairingDistribution("uniform distribution", clientAddr, providersToPair, weightFunc)
}

// Test that random selection of providers is alighned with their stake
func TestPairingDistributionPerStake(t *testing.T) {
	providersCount := 10
	providersToPair := 3

	ts := newTester(t)
	ts.setupForPayments(providersCount, 1, providersToPair)
	_, clientAddr := ts.GetAccount(common.CONSUMER, 0)

	allProviders, err := ts.QueryPairingProviders(ts.spec.Index, false)
	require.NoError(t, err)

	// double the stake of the first provider
	p := allProviders.StakeEntry[0]
	stake := p.Stake.Amount.Int64()
	err = ts.StakeProviderExtra(p.Address, ts.spec, stake*2, p.Endpoints, p.Geolocation, p.Moniker)
	require.NoError(t, err)

	ts.AdvanceEpoch()

	// extend the subscription to accommodate many (pairing) epochs
	_, err = ts.TxSubscriptionBuy(clientAddr, clientAddr, ts.plan.Index, 10)
	require.Nil(t, err)

	weightFunc := func(p epochstoragetypes.StakeEntry) int64 { return p.Stake.Amount.Int64() }
	ts.verifyPairingDistribution("uniform distribution", clientAddr, providersToPair, weightFunc)
}

func unorderedEqual(first, second []string) bool {
	if len(first) != len(second) {
		return false
	}
	exists := make(map[string]bool)
	for _, value := range first {
		exists[value] = true
	}
	for _, value := range second {
		if !exists[value] {
			return false
		}
	}
	return true
}

func IsSubset(subset, superset []string) bool {
	// Create a map to store the elements of the superset
	elements := make(map[string]bool)

	// Populate the map with elements from the superset
	for _, elem := range superset {
		elements[elem] = true
	}

	// Check each element of the subset against the map
	for _, elem := range subset {
		if !elements[elem] {
			return false
		}
	}

	return true
}

func TestGeolocationPairingScores(t *testing.T) {
	ts := newTester(t)
	ts.setupForPayments(1, 3, 1)

	// for convinience
	GL := uint64(planstypes.Geolocation_value["GL"])
	USE := uint64(planstypes.Geolocation_value["USE"])
	EU := uint64(planstypes.Geolocation_value["EU"])
	AS := uint64(planstypes.Geolocation_value["AS"])
	AF := uint64(planstypes.Geolocation_value["AF"])
	AU := uint64(planstypes.Geolocation_value["AU"])
	USC := uint64(planstypes.Geolocation_value["USC"])
	USW := uint64(planstypes.Geolocation_value["USW"])
	USE_EU := USE + EU

	freePlanPolicy := planstypes.Policy{
		GeolocationProfile: 4, // USE
		TotalCuLimit:       10,
		EpochCuLimit:       2,
		MaxProvidersToPair: 5,
	}

	basicPlanPolicy := planstypes.Policy{
		GeolocationProfile: 0, // GLS
		TotalCuLimit:       10,
		EpochCuLimit:       2,
		MaxProvidersToPair: 14,
	}

	premiumPlanPolicy := planstypes.Policy{
		GeolocationProfile: 65535, // GL
		TotalCuLimit:       10,
		EpochCuLimit:       2,
		MaxProvidersToPair: 14,
	}

	// propose all plans and buy subscriptions
	freePlan := planstypes.Plan{
		Index:      "free",
		Block:      ts.BlockHeight(),
		Price:      sdk.NewCoin(epochstoragetypes.TokenDenom, sdk.NewInt(1)),
		PlanPolicy: freePlanPolicy,
	}

	basicPlan := planstypes.Plan{
		Index:      "basic",
		Block:      ts.BlockHeight(),
		Price:      sdk.NewCoin(epochstoragetypes.TokenDenom, sdk.NewInt(1)),
		PlanPolicy: basicPlanPolicy,
	}

	premiumPlan := planstypes.Plan{
		Index:      "premium",
		Block:      ts.BlockHeight(),
		Price:      sdk.NewCoin(epochstoragetypes.TokenDenom, sdk.NewInt(1)),
		PlanPolicy: premiumPlanPolicy,
	}

	plans := []planstypes.Plan{freePlan, basicPlan, premiumPlan}
	err := testkeeper.SimulatePlansAddProposal(ts.Ctx, ts.Keepers.Plans, plans)
	require.Nil(t, err)

	freeAcct, freeAddr := ts.GetAccount(common.CONSUMER, 0)
	basicAcct, basicAddr := ts.GetAccount(common.CONSUMER, 1)
	premiumAcct, premiumAddr := ts.GetAccount(common.CONSUMER, 2)

	ts.TxSubscriptionBuy(freeAddr, freeAddr, freePlan.Index, 1)
	ts.TxSubscriptionBuy(basicAddr, basicAddr, basicPlan.Index, 1)
	ts.TxSubscriptionBuy(premiumAddr, premiumAddr, premiumPlan.Index, 1)

	for geoName, geo := range planstypes.Geolocation_value {
		if geoName != "GL" && geoName != "GLS" {
			err = ts.addProviderGeolocation(5, uint64(geo))
			require.Nil(t, err)
		}
	}

	templates := []struct {
		name         string
		dev          common.Account
		planPolicy   planstypes.Policy
		changePolicy bool
		newGeo       uint64
		expectedGeo  []uint64
	}{
		// free plan (cannot change geolocation - verified in another test)
		{"default free plan", freeAcct, freePlanPolicy, false, 0, []uint64{USE}},

		// basic plan (cannot change geolocation - verified in another test)
		{"default basic plan", basicAcct, basicPlanPolicy, false, 0, []uint64{AF, AS, AU, EU, USE, USC, USW}},

		// premium plan (geolocation can change)
		{"default premium plan", premiumAcct, premiumPlanPolicy, false, 0, []uint64{AF, AS, AU, EU, USE, USC, USW}},
		{"premium plan - set policy regular geo", premiumAcct, premiumPlanPolicy, true, EU, []uint64{EU}},
		{"premium plan - set policy multiple geo", premiumAcct, premiumPlanPolicy, true, USE_EU, []uint64{EU, USE}},
		{"premium plan - set policy global geo", premiumAcct, premiumPlanPolicy, true, GL, []uint64{AF, AS, AU, EU, USE, USC, USW}},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			devResponse, err := ts.QueryProjectDeveloper(tt.dev.Addr.String())
			require.Nil(t, err)

			projIndex := devResponse.Project.Index
			policies := []*planstypes.Policy{&tt.planPolicy}

			newPolicy := planstypes.Policy{}
			if tt.changePolicy {
				newPolicy = tt.planPolicy
				newPolicy.GeolocationProfile = tt.newGeo
				_, err = ts.TxProjectSetPolicy(projIndex, tt.dev.Addr.String(), newPolicy)
				require.Nil(t, err)
				policies = append(policies, &newPolicy)
			}

			ts.AdvanceEpoch() // apply the new policy

			providersRes, err := ts.QueryPairingProviders(ts.spec.Name, false)
			require.Nil(t, err)
			stakeEntries := providersRes.StakeEntry
			providerScores := []*pairingscores.PairingScore{}

			subRes, err := ts.QuerySubscriptionCurrent(tt.dev.Addr.String())
			require.Nil(t, err)
			cluster := subRes.Sub.Cluster

			for i := range stakeEntries {
				qos := ts.Keepers.Pairing.GetQos(ts.Ctx, ts.spec.Index, cluster, stakeEntries[i].Address)
				providerScore := pairingscores.NewPairingScore(&stakeEntries[i], qos)
				providerScores = append(providerScores, providerScore)
			}

			effectiveGeo, err := ts.Keepers.Pairing.CalculateEffectiveGeolocationFromPolicies(policies)
			require.Nil(t, err)

			slots := pairingscores.CalcSlots(&planstypes.Policy{
				GeolocationProfile: effectiveGeo,
				MaxProvidersToPair: tt.planPolicy.MaxProvidersToPair,
			})

			geoSeen := map[uint64]bool{}
			for _, geo := range tt.expectedGeo {
				geoSeen[geo] = false
			}

			// calc scores and verify the scores are as expected
			for _, slot := range slots {
				err = pairingscores.CalcPairingScore(providerScores, pairingscores.GetStrategy(), slot)
				require.Nil(t, err)

				ok := verifyGeoScoreForTesting(providerScores, slot, geoSeen)
				require.True(t, ok)
			}

			// verify that the slots have all the expected geos
			for _, found := range geoSeen {
				require.True(t, found)
			}

			seenIndexes := map[int]struct{}{}
			// check indexes are right
			pairingSlotGroups := pairingscores.GroupSlots(slots)
			for _, pairingSlotGroup := range pairingSlotGroups {
				indexes := pairingSlotGroup.Indexes()
				for _, index := range indexes {
					_, ok := seenIndexes[index]
					require.False(t, ok)
					seenIndexes[index] = struct{}{}
				}
			}
			// verify all slot indexes are in groups
			require.Equal(t, len(seenIndexes), len(slots))
			for idx := range slots {
				_, ok := seenIndexes[idx]
				require.True(t, ok)
			}
		})
	}
}

func isGeoInList(geo uint64, geoList []uint64) bool {
	for _, geoElem := range geoList {
		if geoElem == geo {
			return true
		}
	}
	return false
}

// verifyGeoScoreForTesting is used to testing purposes only!
// it verifies that the max geo score are for providers that exactly match the geo req
// this function assumes that all the other reqs are equal (for example, stake req)
func verifyGeoScoreForTesting(providerScores []*pairingscores.PairingScore, slot *pairingscores.PairingSlot, expectedGeoSeen map[uint64]bool) bool {
	if slot == nil || len(providerScores) == 0 {
		return false
	}

	sort.Slice(providerScores, func(i, j int) bool {
		return providerScores[i].Score.GT(providerScores[j].Score)
	})

	geoReqObject := pairingscores.GeoReq{}
	geoReq, ok := slot.Reqs[geoReqObject.GetName()].(pairingscores.GeoReq)
	if !ok {
		return false
	}

	// verify that the geo is part of the expected geo
	_, ok = expectedGeoSeen[geoReq.Geo]
	if !ok {
		return false
	}
	expectedGeoSeen[geoReq.Geo] = true

	// verify that only providers that match with req geo have max score
	maxScore := providerScores[0].Score
	for _, score := range providerScores {
		if score.Provider.Geolocation == geoReq.Geo {
			if !score.Score.Equal(maxScore) {
				return false
			}
		} else {
			if score.Score.Equal(maxScore) {
				return false
			} else {
				break
			}
		}
	}

	return true
}

func TestDuplicateProviders(t *testing.T) {
	ts := newTester(t)
	ts.setupForPayments(1, 1, 1)

	basicPlanPolicy := planstypes.Policy{
		GeolocationProfile: 0, // GLS
		TotalCuLimit:       10,
		EpochCuLimit:       2,
		MaxProvidersToPair: 14,
	}

	basicPlan := planstypes.Plan{
		Index:      "basic",
		Block:      ts.BlockHeight(),
		Price:      sdk.NewCoin(epochstoragetypes.TokenDenom, sdk.NewInt(1)),
		PlanPolicy: basicPlanPolicy,
	}

	_, basicAddr := ts.GetAccount(common.CONSUMER, 0)

	err := testkeeper.SimulatePlansAddProposal(ts.Ctx, ts.Keepers.Plans, []planstypes.Plan{basicPlan})
	require.Nil(t, err)

	ts.AdvanceEpoch()
	ts.TxSubscriptionBuy(basicAddr, basicAddr, basicPlan.Index, 1)

	for geoName, geo := range planstypes.Geolocation_value {
		if geoName != "GL" && geoName != "GLS" {
			err := ts.addProviderGeolocation(5, uint64(geo))
			require.Nil(t, err)
		}
	}

	ts.AdvanceEpoch()

	for i := 0; i < 100; i++ {
		pairingRes, err := ts.QueryPairingGetPairing(ts.spec.Index, basicAddr)
		require.Nil(t, err)
		providerSeen := map[string]struct{}{}
		for _, provider := range pairingRes.Providers {
			_, found := providerSeen[provider.Address]
			require.False(t, found)
			providerSeen[provider.Address] = struct{}{}
		}
	}
}

// TestNoRequiredGeo checks that if no providers have the required geo, we still get a pairing list
func TestNoRequiredGeo(t *testing.T) {
	ts := newTester(t)
	ts.setupForPayments(1, 1, 5)

	freePlanPolicy := planstypes.Policy{
		GeolocationProfile: 4, // USE
		TotalCuLimit:       10,
		EpochCuLimit:       2,
		MaxProvidersToPair: 5,
	}

	freePlan := planstypes.Plan{
		Index:      "free",
		Block:      ts.BlockHeight(),
		Price:      sdk.NewCoin(epochstoragetypes.TokenDenom, sdk.NewInt(1)),
		PlanPolicy: freePlanPolicy,
	}

	_, freeAddr := ts.GetAccount(common.CONSUMER, 0)

	err := testkeeper.SimulatePlansAddProposal(ts.Ctx, ts.Keepers.Plans, []planstypes.Plan{freePlan})
	require.Nil(t, err)

	ts.AdvanceEpoch()
	ts.TxSubscriptionBuy(freeAddr, freeAddr, freePlan.Index, 1)

	// add 5 more providers that are not in US-E (the only allowed providers in the free plan)
	err = ts.addProviderGeolocation(5, uint64(planstypes.Geolocation_value["AS"]))
	require.Nil(t, err)

	ts.AdvanceEpoch()

	pairingRes, err := ts.QueryPairingGetPairing(ts.spec.Index, freeAddr)
	require.Nil(t, err)
	require.Equal(t, freePlanPolicy.MaxProvidersToPair, uint64(len(pairingRes.Providers)))
	for _, provider := range pairingRes.Providers {
		require.NotEqual(t, freePlanPolicy.GeolocationProfile, provider.Geolocation)
	}
}

// TestGeoSlotCalc checks that the calculated slots always hold a single bit geo req
func TestGeoSlotCalc(t *testing.T) {
	geoReqName := pairingscores.GeoReq{}.GetName()

	allGeos := planstypes.GetAllGeolocations()
	maxGeo := slices.Max(allGeos)

	// iterate over all possible geolocations, create a policy and calc slots
	// not checking 0 because there can never be a policy with geo=0
	for i := 1; i <= int(maxGeo); i++ {
		policy := planstypes.Policy{
			GeolocationProfile: uint64(i),
			MaxProvidersToPair: 14,
		}

		slots := pairingscores.CalcSlots(&policy)
		for _, slot := range slots {
			geoReqFromMap := slot.Reqs[geoReqName]
			geoReq, ok := geoReqFromMap.(pairingscores.GeoReq)
			if !ok {
				require.Fail(t, "slot geo req is not of GeoReq type")
			}

			if !planstypes.IsGeoEnumSingleBit(int32(geoReq.Geo)) {
				require.Fail(t, "slot geo is not single bit")
			}
		}
	}

	// make sure the geo "GL" also works
	policy := planstypes.Policy{
		GeolocationProfile: uint64(planstypes.Geolocation_GL),
		MaxProvidersToPair: 14,
	}
	slots := pairingscores.CalcSlots(&policy)
	for _, slot := range slots {
		geoReqFromMap := slot.Reqs[geoReqName]
		geoReq, ok := geoReqFromMap.(pairingscores.GeoReq)
		if !ok {
			require.Fail(t, "slot geo req is not of GeoReq type")
		}

		if !planstypes.IsGeoEnumSingleBit(int32(geoReq.Geo)) {
			require.Fail(t, "slot geo is not single bit")
		}
	}
}

func TestExtensionAndAddonPairing(t *testing.T) {
	ts := newTester(t)
	ts.setupForPayments(1, 0, 0) // 1 provider, 0 client, default providers-to-pair

	mandatory := spectypes.CollectionData{
		ApiInterface: "mandatory",
		InternalPath: "",
		Type:         "",
		AddOn:        "",
	}
	mandatoryAddon := spectypes.CollectionData{
		ApiInterface: "mandatory",
		InternalPath: "",
		Type:         "",
		AddOn:        "addon",
	}
	optional := spectypes.CollectionData{
		ApiInterface: "optional",
		InternalPath: "",
		Type:         "",
		AddOn:        "optional",
	}
	ts.spec.ApiCollections = []*spectypes.ApiCollection{
		{
			Enabled:        true,
			CollectionData: mandatory,
			Extensions:     getExtensions("ext1", "ext2", "not-supporting-providers"),
		},
		{
			Enabled:        true,
			CollectionData: optional,
			Extensions:     getExtensions("ext1"),
		},
		{
			Enabled:        true,
			CollectionData: mandatoryAddon,
			Extensions:     getExtensions("ext1", "ext2"),
		},
	}

	// will overwrite the default "mock" spec
	ts.AddSpec("mock", ts.spec)
	specId := ts.spec.Index

	mandatoryChainPolicy := &planstypes.ChainPolicy{
		ChainId:      specId,
		Requirements: []planstypes.ChainRequirement{{Collection: mandatory}},
	}
	mandatoryAddonChainPolicy := &planstypes.ChainPolicy{
		ChainId:      specId,
		Requirements: []planstypes.ChainRequirement{{Collection: mandatoryAddon}},
	}
	optionalAddonChainPolicy := &planstypes.ChainPolicy{
		ChainId:      specId,
		Requirements: []planstypes.ChainRequirement{{Collection: optional}},
	}
	optionalAndMandatoryAddonChainPolicy := &planstypes.ChainPolicy{
		ChainId:      specId,
		Requirements: []planstypes.ChainRequirement{{Collection: mandatoryAddon}, {Collection: optional}},
	}
	mandatoryExtChainPolicy := &planstypes.ChainPolicy{
		ChainId: specId,
		Requirements: []planstypes.ChainRequirement{{
			Collection: mandatory,
			Extensions: []string{"ext1"},
		}},
	}
	mandatoryExtChainPolicyMix := &planstypes.ChainPolicy{
		ChainId: specId,
		Requirements: []planstypes.ChainRequirement{{
			Collection: mandatory,
			Extensions: []string{"ext1"},
			Mixed:      true,
		}},
	}
	mandatoryNotSupportingProvidersMix := &planstypes.ChainPolicy{
		ChainId: specId,
		Requirements: []planstypes.ChainRequirement{{
			Collection: mandatory,
			Extensions: []string{"not-supporting-providers"},
			Mixed:      true,
		}},
	}
	mandatoryExt2ChainPolicy := &planstypes.ChainPolicy{
		ChainId: specId,
		Requirements: []planstypes.ChainRequirement{{
			Collection: mandatory,
			Extensions: []string{"ext2"},
		}},
	}
	mandatoryExtBothChainPolicy := &planstypes.ChainPolicy{
		ChainId: specId,
		Requirements: []planstypes.ChainRequirement{{
			Collection: mandatory,
			Extensions: []string{"ext2", "ext1"},
		}},
	}
	mandatoryExtBothSeparatedChainPolicy := &planstypes.ChainPolicy{
		ChainId: specId,
		Requirements: []planstypes.ChainRequirement{
			{
				Collection: mandatory,
				Extensions: []string{"ext2"},
			},
			{
				Collection: mandatory,
				Extensions: []string{"ext1"},
			},
		},
	}
	mandatoryExtAddonChainPolicy := &planstypes.ChainPolicy{
		ChainId:      specId,
		Requirements: []planstypes.ChainRequirement{{Collection: mandatoryAddon, Extensions: []string{"ext1"}}},
	}

	optionalExtAddonChainPolicy := &planstypes.ChainPolicy{
		ChainId:      specId,
		Requirements: []planstypes.ChainRequirement{{Collection: optional, Extensions: []string{"ext1"}}},
	}
	mandatoryExtOptionalChainPolicy := &planstypes.ChainPolicy{
		ChainId:      specId,
		Requirements: []planstypes.ChainRequirement{{Collection: mandatoryAddon, Extensions: []string{"ext1"}}, {Collection: optional}},
	}
	allSupportingChainPolicy := &planstypes.ChainPolicy{
		ChainId:      specId,
		Requirements: []planstypes.ChainRequirement{{Collection: mandatoryAddon, Extensions: []string{"ext1", "ext2"}}, {Collection: optional}},
	}
	templates := []struct {
		name                      string
		planChainPolicy           *planstypes.ChainPolicy
		subscChainPolicy          *planstypes.ChainPolicy
		projChainPolicy           *planstypes.ChainPolicy
		expectedProviders         int
		expectedStrictestPolicies []string
	}{
		{
			name:              "empty",
			expectedProviders: 26,
		},
		{
			name:              "mandatory in plan",
			planChainPolicy:   mandatoryChainPolicy,
			expectedProviders: 25,
		},
		{
			name:              "mandatory in subsc",
			subscChainPolicy:  mandatoryChainPolicy,
			projChainPolicy:   nil,
			expectedProviders: 25,
		},
		{
			name:              "mandatory in proj",
			projChainPolicy:   mandatoryChainPolicy,
			expectedProviders: 25,
		},
		{
			name:                      "addon in plan",
			planChainPolicy:           mandatoryAddonChainPolicy,
			subscChainPolicy:          nil,
			projChainPolicy:           nil,
			expectedProviders:         14,
			expectedStrictestPolicies: []string{"addon"},
		},
		{
			name:                      "addon in subsc",
			subscChainPolicy:          mandatoryAddonChainPolicy,
			expectedProviders:         14,
			expectedStrictestPolicies: []string{"addon"},
		},
		{
			name:                      "addon in proj",
			projChainPolicy:           mandatoryAddonChainPolicy,
			expectedProviders:         14,
			expectedStrictestPolicies: []string{"addon"},
		},
		{
			name:                      "optional in plan",
			planChainPolicy:           optionalAddonChainPolicy,
			expectedProviders:         13,
			expectedStrictestPolicies: []string{"optional"},
		},
		{
			name:                      "optional in subsc",
			subscChainPolicy:          optionalAddonChainPolicy,
			expectedProviders:         13,
			expectedStrictestPolicies: []string{"optional"},
		},
		{
			name:                      "optional in proj",
			projChainPolicy:           optionalAddonChainPolicy,
			expectedProviders:         13,
			expectedStrictestPolicies: []string{"optional"},
		},
		{
			name:                      "optional and addon in plan",
			planChainPolicy:           optionalAndMandatoryAddonChainPolicy,
			expectedProviders:         8,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
		{
			name:                      "optional and addon in subsc",
			subscChainPolicy:          optionalAndMandatoryAddonChainPolicy,
			expectedProviders:         8,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
		{
			name:                      "optional and addon in proj",
			projChainPolicy:           optionalAndMandatoryAddonChainPolicy,
			expectedProviders:         8,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
		{
			name:                      "optional and addon in plan, addon in subsc",
			planChainPolicy:           optionalAndMandatoryAddonChainPolicy,
			subscChainPolicy:          mandatoryAddonChainPolicy,
			expectedProviders:         8,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
		{
			name:                      "optional and addon in subsc, addon in plan",
			planChainPolicy:           mandatoryAddonChainPolicy,
			subscChainPolicy:          optionalAndMandatoryAddonChainPolicy,
			expectedProviders:         8,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
		{
			name:                      "optional and addon in subsc, addon in proj",
			subscChainPolicy:          optionalAndMandatoryAddonChainPolicy,
			projChainPolicy:           mandatoryAddonChainPolicy,
			expectedProviders:         8,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
		{
			name:                      "optional in subsc, addon in proj",
			subscChainPolicy:          optionalAndMandatoryAddonChainPolicy,
			projChainPolicy:           mandatoryAddonChainPolicy,
			expectedProviders:         8,
			expectedStrictestPolicies: []string{"optional", "addon"},
		},
		// check extensions
		{
			name:                      "mandatory ext in plan",
			planChainPolicy:           mandatoryExtChainPolicy,
			expectedProviders:         13,
			expectedStrictestPolicies: []string{"ext1"},
		},
		{
			name:                      "mandatory ext in subsc",
			subscChainPolicy:          mandatoryExtChainPolicy,
			projChainPolicy:           nil,
			expectedProviders:         13,
			expectedStrictestPolicies: []string{"ext1"},
		},
		{
			name:                      "mandatory ext in proj",
			projChainPolicy:           mandatoryExtChainPolicy,
			expectedProviders:         13,
			expectedStrictestPolicies: []string{"ext1"},
		},
		{
			name:                      "mixed mandatory ext in plan",
			planChainPolicy:           mandatoryExtChainPolicyMix,
			expectedProviders:         26,
			expectedStrictestPolicies: []string{"ext1"},
		},
		{
			name:                      "mixed mandatory ext in subsc",
			subscChainPolicy:          mandatoryExtChainPolicyMix,
			projChainPolicy:           nil,
			expectedProviders:         26,
			expectedStrictestPolicies: []string{"ext1"},
		},
		{
			name:                      "mixed mandatory ext in proj",
			projChainPolicy:           mandatoryExtChainPolicyMix,
			expectedProviders:         26,
			expectedStrictestPolicies: []string{"ext1"},
		},
		{
			name:                      "mandatory ext2 in plan",
			planChainPolicy:           mandatoryExt2ChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"ext2"},
		},
		{
			name:                      "mandatory ext2 in subsc",
			subscChainPolicy:          mandatoryExt2ChainPolicy,
			projChainPolicy:           nil,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"ext2"},
		},
		{
			name:                      "mandatory ext2 in proj",
			projChainPolicy:           mandatoryExt2ChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"ext2"},
		},
		{
			name:                      "mandatory ext both in plan",
			planChainPolicy:           mandatoryExtBothChainPolicy,
			expectedProviders:         3,
			expectedStrictestPolicies: []string{"ext1", "ext2"},
		},
		{
			name:                      "mandatory ext both in subsc",
			subscChainPolicy:          mandatoryExtBothChainPolicy,
			projChainPolicy:           nil,
			expectedProviders:         3,
			expectedStrictestPolicies: []string{"ext1", "ext2"},
		},
		{
			name:                      "mandatory ext both in proj",
			projChainPolicy:           mandatoryExtBothChainPolicy,
			expectedProviders:         3,
			expectedStrictestPolicies: []string{"ext1", "ext2"},
		},

		{
			name:                      "mandatory ext both separated in plan",
			planChainPolicy:           mandatoryExtBothSeparatedChainPolicy,
			expectedProviders:         3,
			expectedStrictestPolicies: []string{"ext1", "ext2"},
		},
		{
			name:                      "mandatory ext both separated in subsc",
			subscChainPolicy:          mandatoryExtBothSeparatedChainPolicy,
			projChainPolicy:           nil,
			expectedProviders:         3,
			expectedStrictestPolicies: []string{"ext1", "ext2"},
		},
		{
			name:                      "mandatory ext both separated in proj",
			projChainPolicy:           mandatoryExtBothSeparatedChainPolicy,
			expectedProviders:         3,
			expectedStrictestPolicies: []string{"ext1", "ext2"},
		},
		{
			name:                      "addon ext in plan",
			planChainPolicy:           mandatoryExtAddonChainPolicy,
			subscChainPolicy:          nil,
			projChainPolicy:           nil,
			expectedProviders:         8,
			expectedStrictestPolicies: []string{"addon", "ext1"},
		},
		{
			name:                      "addon ext in subsc",
			subscChainPolicy:          mandatoryExtAddonChainPolicy,
			expectedProviders:         8,
			expectedStrictestPolicies: []string{"addon", "ext1"},
		},
		{
			name:                      "addon ext in proj",
			projChainPolicy:           mandatoryExtAddonChainPolicy,
			expectedProviders:         8,
			expectedStrictestPolicies: []string{"addon", "ext1"},
		},
		{
			name:                      "optional ext in plan",
			planChainPolicy:           optionalExtAddonChainPolicy,
			expectedProviders:         6,
			expectedStrictestPolicies: []string{"optional", "ext1"},
		},
		{
			name:                      "optional ext in subsc",
			subscChainPolicy:          optionalExtAddonChainPolicy,
			expectedProviders:         6,
			expectedStrictestPolicies: []string{"optional", "ext1"},
		},
		{
			name:                      "optional ext in proj",
			projChainPolicy:           optionalExtAddonChainPolicy,
			expectedProviders:         6,
			expectedStrictestPolicies: []string{"optional", "ext1"},
		},
		{
			name:                      "optional ext and addon in plan",
			planChainPolicy:           mandatoryExtOptionalChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon", "ext1"},
		},
		{
			name:                      "optional ext and addon in subsc",
			subscChainPolicy:          mandatoryExtOptionalChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon", "ext1"},
		},
		{
			name:                      "optional ext and addon in proj",
			projChainPolicy:           mandatoryExtOptionalChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon", "ext1"},
		},
		{
			name:                      "optional ext and addon in plan, addon ext in subsc",
			planChainPolicy:           mandatoryExtOptionalChainPolicy,
			subscChainPolicy:          mandatoryExtAddonChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon", "ext1"},
		},
		{
			name:                      "optional ext and addon in subsc, addon ext in plan",
			planChainPolicy:           mandatoryExtAddonChainPolicy,
			subscChainPolicy:          mandatoryExtOptionalChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon", "ext1"},
		},
		{
			name:                      "optional ext and addon in subsc, addon ext in proj",
			subscChainPolicy:          mandatoryExtOptionalChainPolicy,
			projChainPolicy:           mandatoryExtAddonChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon", "ext1"},
		},
		{
			name:                      "optional ext in subsc, addon ext in proj",
			subscChainPolicy:          mandatoryExtOptionalChainPolicy,
			projChainPolicy:           mandatoryExtAddonChainPolicy,
			expectedProviders:         4,
			expectedStrictestPolicies: []string{"optional", "addon", "ext1"},
		},
		{
			name:                      "all supporting in plan",
			planChainPolicy:           allSupportingChainPolicy,
			expectedProviders:         2,
			expectedStrictestPolicies: []string{"optional", "addon", "ext1", "ext2"},
		},
		{
			name:                      "all supporting in subsc",
			subscChainPolicy:          allSupportingChainPolicy,
			projChainPolicy:           nil,
			expectedProviders:         2,
			expectedStrictestPolicies: []string{"optional", "addon", "ext1", "ext2"},
		},
		{
			name:                      "all supporting in proj",
			projChainPolicy:           allSupportingChainPolicy,
			expectedProviders:         2,
			expectedStrictestPolicies: []string{"optional", "addon", "ext1", "ext2"},
		},
		{
			name:              "mixed not supporting providers",
			projChainPolicy:   mandatoryNotSupportingProvidersMix,
			expectedProviders: 26,
		},
	}

	mandatorySupportingEndpoints := []epochstoragetypes.Endpoint{{
		IPPORT:        "123",
		Geolocation:   1,
		Addons:        []string{mandatory.AddOn},
		ApiInterfaces: []string{mandatory.ApiInterface},
	}}
	mandatoryAddonSupportingEndpoints := []epochstoragetypes.Endpoint{{
		IPPORT:        "456",
		Geolocation:   1,
		Addons:        []string{mandatoryAddon.AddOn},
		ApiInterfaces: []string{mandatoryAddon.ApiInterface},
	}}
	mandatoryAndMandatoryAddonSupportingEndpoints := slices.Concat(
		mandatorySupportingEndpoints, mandatoryAddonSupportingEndpoints)

	optionalSupportingEndpoints := []epochstoragetypes.Endpoint{{
		IPPORT:        "789",
		Geolocation:   1,
		Addons:        []string{optional.AddOn},
		ApiInterfaces: []string{optional.ApiInterface},
	}}
	optionalAndMandatorySupportingEndpoints := slices.Concat(
		mandatorySupportingEndpoints, optionalSupportingEndpoints)
	optionalAndMandatoryAddonSupportingEndpoints := slices.Concat(
		mandatoryAddonSupportingEndpoints, optionalSupportingEndpoints)

	allSupportingEndpoints := slices.Concat(
		mandatorySupportingEndpoints, optionalAndMandatoryAddonSupportingEndpoints)

	mandatoryAndOptionalSingleEndpoint := []epochstoragetypes.Endpoint{{
		IPPORT:        "444",
		Geolocation:   1,
		Addons:        []string{},
		ApiInterfaces: []string{mandatoryAddon.ApiInterface, optional.ApiInterface},
	}}

	// now with extensions

	mandatoryExtSupportingEndpoints := []epochstoragetypes.Endpoint{{
		IPPORT:        "Ext-1",
		Geolocation:   1,
		Addons:        []string{mandatory.AddOn},
		ApiInterfaces: []string{mandatory.ApiInterface},
		Extensions:    []string{"ext1"},
	}}

	mandatoryExt2SupportingEndpoint := []epochstoragetypes.Endpoint{{
		IPPORT:        "Ext-2",
		Geolocation:   1,
		Addons:        []string{mandatory.AddOn},
		ApiInterfaces: []string{mandatory.ApiInterface},
		Extensions:    []string{"ext2"},
	}}

	mandatoryExt2AddonSupportingEndpoint := []epochstoragetypes.Endpoint{{
		IPPORT:        "Ext-2-addon",
		Geolocation:   1,
		Addons:        []string{mandatoryAddon.AddOn},
		ApiInterfaces: []string{mandatoryAddon.ApiInterface},
		Extensions:    []string{"ext2"},
	}}

	mandatoryExtBOTHSupportingEndpoint := []epochstoragetypes.Endpoint{{
		IPPORT:        "Ext-both",
		Geolocation:   1,
		Addons:        []string{mandatory.AddOn},
		ApiInterfaces: []string{mandatory.ApiInterface},
		Extensions:    []string{"ext1", "ext2"},
	}}

	mandatoryExtAddonSupportingEndpoints := []epochstoragetypes.Endpoint{{
		IPPORT:        "Ext-3",
		Geolocation:   1,
		Addons:        []string{mandatoryAddon.AddOn},
		ApiInterfaces: []string{mandatoryAddon.ApiInterface},
		Extensions:    []string{"ext1"},
	}}
	mandatoryExtAndMandatoryAddonSupportingEndpoints := slices.Concat(
		mandatoryExtSupportingEndpoints, mandatoryExtAddonSupportingEndpoints)

	optionalExtSupportingEndpoints := []epochstoragetypes.Endpoint{{
		IPPORT:        "Ext-4",
		Geolocation:   1,
		Addons:        []string{optional.AddOn},
		ApiInterfaces: []string{optional.ApiInterface},
		Extensions:    []string{"ext1"},
	}}
	optionalExtAndMandatorySupportingEndpoints := slices.Concat(
		mandatoryExtSupportingEndpoints, optionalExtSupportingEndpoints)
	optionalExtAndMandatoryAddonSupportingEndpoints := slices.Concat(
		mandatoryExtAddonSupportingEndpoints, optionalExtSupportingEndpoints)

	allExtSupportingEndpoints := slices.Concat(
		mandatoryExtSupportingEndpoints, optionalExtAndMandatoryAddonSupportingEndpoints, mandatoryExt2AddonSupportingEndpoint)
	// mandatory
	err := ts.addProviderEndpoints(2, mandatoryExtSupportingEndpoints) // ext1 - 2
	require.NoError(t, err)
	err = ts.addProviderEndpoints(2, mandatorySupportingEndpoints)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(1, mandatoryExt2SupportingEndpoint) // ext2 - 1
	require.NoError(t, err)
	err = ts.addProviderEndpoints(1, mandatoryExtBOTHSupportingEndpoint) // ext1 + ext2 - 1
	require.NoError(t, err)
	// mandatory + addon
	err = ts.addProviderEndpoints(2, mandatoryAndMandatoryAddonSupportingEndpoints)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(2, mandatoryExtAddonSupportingEndpoints)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(2, mandatoryExtAndMandatoryAddonSupportingEndpoints)
	require.NoError(t, err)
	// mandatory + optional
	err = ts.addProviderEndpoints(2, optionalAndMandatorySupportingEndpoints)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(1, mandatoryAndOptionalSingleEndpoint)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(2, optionalExtAndMandatorySupportingEndpoints)
	require.NoError(t, err)
	// mandatory + optional + addon
	err = ts.addProviderEndpoints(2, optionalAndMandatoryAddonSupportingEndpoints)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(2, allSupportingEndpoints)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(2, optionalExtAndMandatoryAddonSupportingEndpoints)
	require.NoError(t, err)
	err = ts.addProviderEndpoints(2, allExtSupportingEndpoints)
	require.NoError(t, err)

	// summary of endpoints:
	// ext1 has 13 supporting providers
	// ext 2 has 4 supporting providers
	// addons have  14
	// optional has 13
	// addon + optional has 8
	// addon + ext has 8
	// optional + ext has 6
	// all supporting one ext (ext1 addon optional) has 4
	// all supporting (ext2 ext1 addon optional) has 2

	// erroring out
	err = ts.addProviderEndpoints(2, optionalSupportingEndpoints) // this errors as it doesnt implement mandatory
	require.Error(t, err)
	err = ts.addProviderEndpoints(2, optionalExtSupportingEndpoints) // this errors as it doesnt implement mandatory
	require.Error(t, err)

	stakeStorage, found := ts.Keepers.Epochstorage.GetStakeStorageCurrent(ts.Ctx, ts.spec.Index)
	require.True(t, found)
	require.Len(t, stakeStorage.StakeEntries, 26) // one for stub and 25 others

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			defaultPolicy := func() planstypes.Policy {
				return planstypes.Policy{
					ChainPolicies:      []planstypes.ChainPolicy{},
					GeolocationProfile: math.MaxUint64,
					MaxProvidersToPair: 100,
					TotalCuLimit:       math.MaxUint64,
					EpochCuLimit:       math.MaxUint64,
				}
			}

			plan := ts.plan // original mock template
			plan.PlanPolicy = defaultPolicy()

			if tt.planChainPolicy != nil {
				plan.PlanPolicy.ChainPolicies = []planstypes.ChainPolicy{*tt.planChainPolicy}
			}

			err := ts.TxProposalAddPlans(plan)
			require.Nil(t, err)

			_, sub1Addr := ts.AddAccount("sub", 0, 10000)

			_, err = ts.TxSubscriptionBuy(sub1Addr, sub1Addr, plan.Index, 1)
			require.Nil(t, err)

			// get the admin project and set its policies
			subProjects, err := ts.QuerySubscriptionListProjects(sub1Addr)
			require.Nil(t, err)
			require.Equal(t, 1, len(subProjects.Projects))

			projectID := subProjects.Projects[0]

			if tt.projChainPolicy != nil {
				projPolicy := defaultPolicy()
				projPolicy.ChainPolicies = []planstypes.ChainPolicy{*tt.projChainPolicy}
				_, err = ts.TxProjectSetPolicy(projectID, sub1Addr, projPolicy)
				require.Nil(t, err)
			}

			// apply policy change
			ts.AdvanceEpoch()

			if tt.subscChainPolicy != nil {
				subscPolicy := defaultPolicy()
				subscPolicy.ChainPolicies = []planstypes.ChainPolicy{*tt.subscChainPolicy}
				_, err = ts.TxProjectSetSubscriptionPolicy(projectID, sub1Addr, subscPolicy)
				require.Nil(t, err)
			}

			// apply policy change
			ts.AdvanceEpochs(2)

			project, err := ts.GetProjectForBlock(projectID, ts.BlockHeight())
			require.NoError(t, err)

			strictestPolicy, _, err := ts.Keepers.Pairing.GetProjectStrictestPolicy(ts.Ctx, project, specId)
			require.NoError(t, err)
			if len(tt.expectedStrictestPolicies) > 0 {
				require.NotEqual(t, 0, len(strictestPolicy.ChainPolicies))
				require.NotEqual(t, 0, len(strictestPolicy.ChainPolicies[0].Requirements))
				services := map[string]struct{}{}
				for _, requirement := range strictestPolicy.ChainPolicies[0].Requirements {
					collection := requirement.Collection
					if collection.AddOn != "" {
						services[collection.AddOn] = struct{}{}
					}
					for _, extension := range requirement.Extensions {
						if extension != "" {
							services[extension] = struct{}{}
						}
					}
				}
				for _, expected := range tt.expectedStrictestPolicies {
					_, ok := services[expected]
					require.True(t, ok, "did not find addon in strictest policy %s, policy: %#v", expected, strictestPolicy)
				}
			}

			pairing, err := ts.QueryPairingGetPairing(ts.spec.Index, sub1Addr)
			if tt.expectedProviders > 0 {
				require.Nil(t, err)
				require.Equal(t, tt.expectedProviders, len(pairing.Providers), "received providers %#v", pairing)
				if len(tt.expectedStrictestPolicies) > 0 {
					services := map[string]int{}
					for _, provider := range pairing.GetProviders() {
						for _, endpoint := range provider.Endpoints {
							for _, addon := range endpoint.Addons {
								services[addon]++
							}
							for _, extension := range endpoint.Extensions {
								services[extension]++
							}
							for _, apiInterface := range endpoint.ApiInterfaces {
								services[apiInterface]++
							}
						}
					}
					for _, expected := range tt.expectedStrictestPolicies {
						count, ok := services[expected]
						require.True(t, ok, "did not find addon in strictest policy %s, policy: %#v", expected, services)
						require.GreaterOrEqual(t, count, len(pairing.Providers)/2) // we expect at least half of the providers to support the expected api interface (for mix it's half)
					}
				}
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestMixSelectedProvidersAndArchivePairing(t *testing.T) {
	ts := newTester(t)
	ts.setupForPayments(1, 0, 0) // 1 provider, 0 client, default providers-to-pair
	specEth, err := testkeeper.GetASpec("ETH1", "../../../", nil, nil)
	if err != nil {
		require.NoError(t, err)
	}
	ts.spec.ApiCollections = specEth.ApiCollections

	// will overwrite the default "mock" spec
	ts.AddSpec("mock", ts.spec)
	specId := ts.spec.Index
	mandatoryExtChainPolicyMix := planstypes.ChainPolicy{
		ChainId: specId,
		Requirements: []planstypes.ChainRequirement{{
			Collection: spectypes.CollectionData{
				ApiInterface: "jsonrpc",
				InternalPath: "",
				Type:         "POST",
				AddOn:        "",
			},
			Extensions: []string{"archive"},
			Mixed:      true,
		}},
	}

	regularEndpoints := []epochstoragetypes.Endpoint{{
		IPPORT:        "123",
		Geolocation:   1,
		Addons:        []string{},
		ApiInterfaces: []string{},
	}}
	archiveEndpoints := []epochstoragetypes.Endpoint{{
		IPPORT:        "123",
		Geolocation:   1,
		Addons:        []string{"archive"},
		ApiInterfaces: []string{},
	}}

	// mandatory
	err = ts.addProviderEndpoints(200, regularEndpoints) // ext1 - 2
	require.NoError(t, err)

	_, p1 := ts.GetAccount(common.PROVIDER, 0)
	_, p2 := ts.GetAccount(common.PROVIDER, 1)
	_, p3 := ts.GetAccount(common.PROVIDER, 2)
	_, p4 := ts.GetAccount(common.PROVIDER, 3)
	_, p5 := ts.GetAccount(common.PROVIDER, 4)
	selectedProviders := []string{p1, p2, p3, p4, p5}
	err = ts.addProviderEndpoints(10, archiveEndpoints) // ext1 - 2
	require.NoError(t, err)

	t.Run("archive selected providers mixed test", func(t *testing.T) {
		defaultPolicy := func() planstypes.Policy {
			return planstypes.Policy{
				ChainPolicies:      []planstypes.ChainPolicy{},
				GeolocationProfile: math.MaxUint64,
				MaxProvidersToPair: 30,
				TotalCuLimit:       math.MaxUint64,
				EpochCuLimit:       math.MaxUint64,
			}
		}

		plan := ts.plan // original mock template
		plan.PlanPolicy = defaultPolicy()
		plan.PlanPolicy.SelectedProvidersMode = planstypes.SELECTED_PROVIDERS_MODE_MIXED
		plan.PlanPolicy.SelectedProviders = selectedProviders

		plan.PlanPolicy.ChainPolicies = []planstypes.ChainPolicy{mandatoryExtChainPolicyMix}

		expectedProviders := plan.PlanPolicy.MaxProvidersToPair

		err := ts.TxProposalAddPlans(plan)
		require.Nil(t, err)

		_, sub1Addr := ts.AddAccount("sub", 0, 10000)

		_, err = ts.TxSubscriptionBuy(sub1Addr, sub1Addr, plan.Index, 1)
		require.Nil(t, err)

		// get the admin project and set its policies
		subProjects, err := ts.QuerySubscriptionListProjects(sub1Addr)
		require.Nil(t, err)
		require.Equal(t, 1, len(subProjects.Projects))

		projectID := subProjects.Projects[0]

		// apply policy change
		ts.AdvanceEpoch()

		// apply policy change
		ts.AdvanceEpochs(2)

		project, err := ts.GetProjectForBlock(projectID, ts.BlockHeight())
		require.NoError(t, err)

		strictestPolicy, _, err := ts.Keepers.Pairing.GetProjectStrictestPolicy(ts.Ctx, project, specId)
		require.NoError(t, err)

		require.NotEqual(t, 0, len(strictestPolicy.ChainPolicies))
		require.NotEqual(t, 0, len(strictestPolicy.ChainPolicies[0].Requirements))
		services := map[string]struct{}{}
		for _, requirement := range strictestPolicy.ChainPolicies[0].Requirements {
			collection := requirement.Collection
			if collection.AddOn != "" {
				services[collection.AddOn] = struct{}{}
			}
			for _, extension := range requirement.Extensions {
				if extension != "" {
					services[extension] = struct{}{}
				}
			}
		}
		for _, expected := range []string{"archive"} {
			_, ok := services[expected]
			require.True(t, ok, "did not find addon in strictest policy %s, policy: %#v", expected, strictestPolicy)
		}

		pairing, err := ts.QueryPairingGetPairing(ts.spec.Index, sub1Addr)
		require.Nil(t, err)
		require.Equal(t, expectedProviders, uint64(len(pairing.Providers)), "received providers %#v", pairing)

		servicesCount := map[string]int{}
		for _, provider := range pairing.GetProviders() {
			for _, endpoint := range provider.Endpoints {
				for _, addon := range endpoint.Addons {
					servicesCount[addon]++
				}
				for _, extension := range endpoint.Extensions {
					servicesCount[extension]++
				}
				for _, apiInterface := range endpoint.ApiInterfaces {
					servicesCount[apiInterface]++
				}
			}
		}
		for _, expected := range []string{"archive"} {
			count, ok := servicesCount[expected]
			require.True(t, ok, "did not find addon in strictest policy %s, policy: %#v", expected, services)
			require.GreaterOrEqual(t, count, len(pairing.Providers)/3) // we expect at least third of the providers to support the expected api interface
		}
		// verify selected providers mix count
		addresses := []string{}
		for _, provider := range pairing.Providers {
			addresses = append(addresses, provider.Address)
		}
		count := countSelectedAddresses(addresses, selectedProviders)
		require.Equal(t, count, len(selectedProviders))
	})
}

// TestPairingConsistency checks we consistently get the same pairing in the same epoch
func TestPairingConsistency(t *testing.T) {
	ts := newTester(t)
	iterations := 100

	ts.plan.PlanPolicy.MaxProvidersToPair = uint64(3)
	ts.AddPlan("mock", ts.plan)
	ts.addClient(1)
	err := ts.addProviderGeolocation(10, 3)
	require.Nil(t, err)

	ts.AdvanceEpoch()

	consumers := ts.Accounts(common.CONSUMER)

	res, err := ts.QueryPairingGetPairing(ts.spec.Index, consumers[0].Addr.String())
	require.Nil(t, err)
	prevPairing := res.Providers
	for i := 0; i < iterations; i++ {
		res, err := ts.QueryPairingGetPairing(ts.spec.Index, consumers[0].Addr.String())
		require.Nil(t, err)

		var prevPairingAddrs []string
		var currentPairingAddrs []string

		for i := range res.Providers {
			prevPairingAddrs = append(prevPairingAddrs, prevPairing[i].Address)
			currentPairingAddrs = append(currentPairingAddrs, res.Providers[i].Address)
		}

		require.True(t, slices.UnorderedEqual(prevPairingAddrs, currentPairingAddrs))

		prevPairing = res.Providers
	}
}

// TestNoZeroLatency checks that there are no zero values in GEO_LATENCY_MAP
func TestNoZeroLatency(t *testing.T) {
	for _, latencyMap := range pairingscores.GEO_LATENCY_MAP {
		for _, latency := range latencyMap {
			require.NotEqual(t, uint64(0), latency)
		}
	}
}
