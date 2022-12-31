// Copyright 2022 Piprate Limited
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package operations

import (
	"fmt"

	"github.com/piprate/metalocker/model"
	"github.com/piprate/metalocker/wallet"
	"github.com/xlab/treeprint"
)

func idWithLevel(lvl model.AccessLevel, id string) string {
	switch lvl {
	case model.AccessLevelNone:
		return " 0 " + id + " "
	case model.AccessLevelRestricted:
		return " R " + id + " "
	case model.AccessLevelManaged:
		return " M " + id + " "
	case model.AccessLevelHosted:
		return " H " + id + " "
	case model.AccessLevelLocal:
		return " L " + id + " "
	case model.AccessLevelCold:
		return " C " + id + " "
	}
	return " ? " + id + " "
}

func printAccount(dw wallet.DataWallet, tree treeprint.Tree) error {
	tree = tree.AddMetaBranch(idWithLevel(dw.Account().AccessLevel, dw.ID()), dw.Account().Name)

	subs, err := dw.SubAccounts()
	if err != nil {
		return err
	}

	if len(subs) > 0 {
		subTree := tree.AddBranch("sub-accounts")
		for _, sub := range subs {
			subDW, err := dw.GetSubAccountWallet(sub.ID)
			if err != nil {
				return err
			}
			defer subDW.Close()
			if err = printAccount(subDW, subTree); err != nil {
				return err
			}
			_ = subDW.Close()
		}
	}

	idList, err := dw.GetIdentities()
	if err != nil {
		return err
	}

	lockers, err := dw.GetLockers()
	if err != nil {
		return err
	}

	lockerMap := make(map[string][]*model.Locker)
	var orphanLockers []*model.Locker
	for _, l := range lockers {
		us := l.Us()
		if us != nil {
			lockerMap[us.ID] = append(lockerMap[us.ID], l)
		} else {
			orphanLockers = append(orphanLockers, l)
		}
	}

	if len(idList) > 0 {
		idyTree := tree.AddBranch("identities")
		for _, idy := range idList {
			idyBranch := idyTree.AddMetaBranch(idWithLevel(idy.AccessLevel(), idy.ID()), idy.Name())
			for _, l := range lockerMap[idy.ID()] {
				idyBranch.AddMetaNode(l.ID, l.Name)
			}

		}
	}
	if len(orphanLockers) > 0 {
		lockerTree := tree.AddBranch("lockers")
		for _, l := range orphanLockers {
			lockerTree.AddMetaNode(idWithLevel(l.AccessLevel, l.ID), l.Name)
		}
	}

	return nil
}

func ChartToString(dw wallet.DataWallet, name string) (string, error) {
	tree := treeprint.New()
	tree.SetValue(name)

	if err := printAccount(dw, tree); err != nil {
		return "", err
	}

	return tree.String(), nil
}

func PrintWallet(dw wallet.DataWallet, name string) error {
	tree := treeprint.New()
	tree.SetValue(name)

	if err := printAccount(dw, tree); err != nil {
		return err
	}

	val, err := ChartToString(dw, name)
	if err != nil {
		return err
	}

	fmt.Println(val)

	return nil
}
