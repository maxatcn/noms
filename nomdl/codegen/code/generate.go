// Package code provides Generator, which has methods for generating code snippets from a types.TypeRef.
// Conceptually there are few type spaces here:
//
// - Def - MyStructDef, ListOfBoolDef; convenient Go types for working with data from a given Noms Value.
// - Native - such as string, uint32
// - Value - the generic types.Value
// - Nom - types.String, types.UInt32, MyStruct, ListOfBool
// - User - User defined structs, enums etc as well as native primitves. This uses Native when possible or Nom if not. These are to be used in APIs for generated types -- Getters and setters for maps and structs, etc.
package code

import (
	"fmt"
	"strings"

	"github.com/attic-labs/noms/d"
	"github.com/attic-labs/noms/ref"
	"github.com/attic-labs/noms/types"
)

// Resolver provides a single method for resolving an unresolved types.TypeRef.
type Resolver interface {
	Resolve(t types.TypeRef) types.TypeRef
}

// Generator provides methods for generating code snippets from both resolved and unresolved types.TypeRefs. In the latter case, it uses R to resolve the types.TypeRef before generating code.
type Generator struct {
	R Resolver
}

// DefType returns a string containing the Go type that should be used as the 'Def' for the Noms type described by t.
func (gen Generator) DefType(t types.TypeRef) string {
	rt := gen.R.Resolve(t)
	k := rt.Kind()
	switch k {
	case types.BlobKind:
		return "types.Blob"
	case types.BoolKind, types.Float32Kind, types.Float64Kind, types.Int16Kind, types.Int32Kind, types.Int64Kind, types.Int8Kind, types.StringKind, types.UInt16Kind, types.UInt32Kind, types.UInt64Kind, types.UInt8Kind:
		return strings.ToLower(kindToString(k))
	case types.EnumKind:
		return gen.UserName(t)
	case types.ListKind, types.MapKind, types.SetKind, types.StructKind:
		return gen.UserName(t) + "Def"
	case types.PackageKind:
		return "types.Package"
	case types.RefKind:
		return "ref.Ref"
	case types.ValueKind:
		return "types.Value"
	case types.TypeRefKind:
		return "types.TypeRef"
	}
	panic("unreachable")
}

// UserType returns a string containing the Go type that should be used when the Noms type described by t needs to be returned by a generated getter or taken as a parameter to a generated setter.
func (gen Generator) UserType(t types.TypeRef) string {
	rt := gen.R.Resolve(t)
	k := rt.Kind()
	switch k {
	case types.BlobKind:
		return "types.Blob"
	case types.BoolKind, types.Float32Kind, types.Float64Kind, types.Int16Kind, types.Int32Kind, types.Int64Kind, types.Int8Kind, types.StringKind, types.UInt16Kind, types.UInt32Kind, types.UInt64Kind, types.UInt8Kind:
		return strings.ToLower(kindToString(k))
	case types.EnumKind, types.ListKind, types.MapKind, types.RefKind, types.SetKind, types.StructKind:
		return gen.UserName(t)
	case types.PackageKind:
		return "types.Package"
	case types.ValueKind:
		return "types.Value"
	case types.TypeRefKind:
		return "types.TypeRef"
	}
	panic("unreachable")
}

// DefToValue returns a string containing Go code to convert an instance of a Def type (named val) to a Noms types.Value of the type described by t.
func (gen Generator) DefToValue(val string, t types.TypeRef) string {
	rt := gen.R.Resolve(t)
	switch rt.Kind() {
	case types.BlobKind, types.EnumKind, types.PackageKind, types.ValueKind, types.TypeRefKind:
		return val // No special Def representation
	case types.BoolKind, types.Float32Kind, types.Float64Kind, types.Int16Kind, types.Int32Kind, types.Int64Kind, types.Int8Kind, types.StringKind, types.UInt16Kind, types.UInt32Kind, types.UInt64Kind, types.UInt8Kind:
		return gen.NativeToValue(val, rt)
	case types.ListKind, types.MapKind, types.SetKind, types.StructKind:
		return fmt.Sprintf("%s.New()", val)
	case types.RefKind:
		return fmt.Sprintf("New%s(%s)", gen.UserName(rt), val)
	}
	panic("unreachable")
}

// ValueToDef returns a string containing Go code to convert an instance of a types.Value (val) into the Def type appropriate for t.
func (gen Generator) ValueToDef(val string, t types.TypeRef) string {
	rt := gen.R.Resolve(t)
	switch rt.Kind() {
	case types.BlobKind, types.PackageKind, types.TypeRefKind:
		return gen.ValueToUser(val, rt) // No special Def representation
	case types.BoolKind, types.Float32Kind, types.Float64Kind, types.Int16Kind, types.Int32Kind, types.Int64Kind, types.Int8Kind, types.StringKind, types.UInt16Kind, types.UInt32Kind, types.UInt64Kind, types.UInt8Kind:
		return gen.ValueToNative(val, rt)
	case types.EnumKind:
		return fmt.Sprintf("%s.(%s)", val, gen.UserName(t))
	case types.ListKind, types.MapKind, types.SetKind, types.StructKind:
		return fmt.Sprintf("%s.Def()", gen.ValueToUser(val, t))
	case types.RefKind:
		return fmt.Sprintf("%s.Ref()", val)
	case types.ValueKind:
		return val // Value is already a Value
	}
	panic("unreachable")
}

// NativeToValue returns a string containing Go code to convert an instance of a native type (named val) to a Noms types.Value of the type described by t.
func (gen Generator) NativeToValue(val string, t types.TypeRef) string {
	t = gen.R.Resolve(t)
	k := t.Kind()
	switch k {
	case types.BoolKind, types.Float32Kind, types.Float64Kind, types.Int16Kind, types.Int32Kind, types.Int64Kind, types.Int8Kind, types.UInt16Kind, types.UInt32Kind, types.UInt64Kind, types.UInt8Kind:
		return fmt.Sprintf("types.%s(%s)", kindToString(k), val)
	case types.StringKind:
		return "types.NewString(" + val + ")"
	}
	panic("unreachable")
}

// ValueToNative returns a string containing Go code to convert an instance of a types.Value (val) into the native type appropriate for t.
func (gen Generator) ValueToNative(val string, t types.TypeRef) string {
	k := t.Kind()
	switch k {
	case types.BoolKind, types.Float32Kind, types.Float64Kind, types.Int16Kind, types.Int32Kind, types.Int64Kind, types.Int8Kind, types.UInt16Kind, types.UInt32Kind, types.UInt64Kind, types.UInt8Kind:
		n := kindToString(k)
		return fmt.Sprintf("%s(%s.(types.%s))", strings.ToLower(n), val, n)
	case types.StringKind:
		return val + ".(types.String).String()"
	}
	panic("unreachable")
}

// UserToValue returns a string containing Go code to convert an instance of a User type (named val) to a Noms types.Value of the type described by t. For Go primitive types, this will use NativeToValue(). For other types, their UserType is a Noms types.Value (or a wrapper around one), so this is more-or-less a pass-through.
func (gen Generator) UserToValue(val string, t types.TypeRef) string {
	t = gen.R.Resolve(t)
	k := t.Kind()
	switch k {
	case types.BlobKind, types.EnumKind, types.ListKind, types.MapKind, types.PackageKind, types.RefKind, types.SetKind, types.StructKind, types.TypeRefKind, types.ValueKind:
		return val
	case types.BoolKind, types.Float32Kind, types.Float64Kind, types.Int16Kind, types.Int32Kind, types.Int64Kind, types.Int8Kind, types.StringKind, types.UInt16Kind, types.UInt32Kind, types.UInt64Kind, types.UInt8Kind:
		return gen.NativeToValue(val, t)
	}
	panic("unreachable")
}

// ValueToUser returns a string containing Go code to convert an instance of a types.Value (val) into the User type appropriate for t. For Go primitives, this will use ValueToNative().
func (gen Generator) ValueToUser(val string, t types.TypeRef) string {
	rt := gen.R.Resolve(t)
	k := rt.Kind()
	switch k {
	case types.BlobKind:
		return fmt.Sprintf("%s.(types.Blob)", val)
	case types.BoolKind, types.Float32Kind, types.Float64Kind, types.Int16Kind, types.Int32Kind, types.Int64Kind, types.Int8Kind, types.StringKind, types.UInt16Kind, types.UInt32Kind, types.UInt64Kind, types.UInt8Kind:
		return gen.ValueToNative(val, rt)
	case types.EnumKind, types.ListKind, types.MapKind, types.RefKind, types.SetKind, types.StructKind:
		return fmt.Sprintf("%s.(%s)", val, gen.UserName(t))
	case types.PackageKind:
		return fmt.Sprintf("%s.(types.Package)", val)
	case types.ValueKind:
		return val
	case types.TypeRefKind:
		return fmt.Sprintf("%s.(types.TypeRef)", val)
	}
	panic("unreachable")
}

// UserZero returns a string containing Go code to create an uninitialized instance of the User type appropriate for t.
func (gen Generator) UserZero(t types.TypeRef) string {
	rt := gen.R.Resolve(t)
	k := rt.Kind()
	switch k {
	case types.BlobKind:
		return "types.NewEmptyBlob()"
	case types.BoolKind:
		return "false"
	case types.EnumKind:
		return gen.newUserName(t, rt)
	case types.Float32Kind, types.Float64Kind, types.Int16Kind, types.Int32Kind, types.Int64Kind, types.Int8Kind, types.UInt16Kind, types.UInt32Kind, types.UInt64Kind, types.UInt8Kind:
		return fmt.Sprintf("%s(0)", strings.ToLower(kindToString(k)))
	case types.ListKind, types.MapKind, types.PackageKind, types.SetKind, types.StructKind:
		return gen.newUserName(t, rt)
	case types.RefKind:
		return fmt.Sprintf("New%s(ref.Ref{})", gen.UserName(rt))
	case types.StringKind:
		return `""`
	case types.ValueKind:
		// TODO: This is where a null Value would have been useful.
		return "types.Bool(false)"
	case types.TypeRefKind:
		return "types.TypeRef{R: ref.Ref{}}"
	}
	panic("unreachable")
}

func (gen Generator) newUserName(t, rt types.TypeRef) string {
	if t.HasPackageRef() {
		return fmt.Sprintf("%s.New%s()", ToTag(t.PackageRef()), gen.UserName(rt))
	}
	return fmt.Sprintf("New%s()", gen.UserName(rt))
}

// ValueZero returns a string containing Go code to create an uninitialized instance of the Noms types.Value appropriate for t.
func (gen Generator) ValueZero(t types.TypeRef) string {
	rt := gen.R.Resolve(t)
	k := rt.Kind()
	switch k {
	case types.BlobKind:
		return "types.NewEmptyBlob()"
	case types.BoolKind:
		return "types.Bool(false)"
	case types.EnumKind:
		return gen.UserZero(t)
	case types.Float32Kind, types.Float64Kind, types.Int16Kind, types.Int32Kind, types.Int64Kind, types.Int8Kind, types.UInt16Kind, types.UInt32Kind, types.UInt64Kind, types.UInt8Kind:
		return fmt.Sprintf("types.%s(0)", kindToString(k))
	case types.ListKind, types.MapKind, types.RefKind, types.SetKind:
		return gen.UserZero(t)
	case types.PackageKind:
		return "types.NewPackage()"
	case types.StringKind:
		return `types.NewString("")`
	case types.StructKind:
		return gen.newUserName(t, rt)
	case types.ValueKind:
		// TODO: Use nil here
		return "types.Bool(false)"
	case types.TypeRefKind:
		return "types.TypeRef{}"
	}
	panic("unreachable")
}

// UserName returns the name of the User type appropriate for t, taking into account Noms types imported from other packages.
func (gen Generator) UserName(t types.TypeRef) string {
	rt := gen.R.Resolve(t)
	k := rt.Kind()
	switch k {
	case types.BlobKind, types.BoolKind, types.Float32Kind, types.Float64Kind, types.Int16Kind, types.Int32Kind, types.Int64Kind, types.Int8Kind, types.PackageKind, types.StringKind, types.UInt16Kind, types.UInt32Kind, types.UInt64Kind, types.UInt8Kind, types.ValueKind, types.TypeRefKind:
		return kindToString(k)
	case types.EnumKind:
		if t.HasPackageRef() {
			return ToTag(t.PackageRef()) + "." + rt.Name()
		}
		return rt.Name()
	case types.ListKind:
		return fmt.Sprintf("ListOf%s", gen.refToId(rt.Desc.(types.CompoundDesc).ElemTypes[0]))
	case types.MapKind:
		elemTypes := rt.Desc.(types.CompoundDesc).ElemTypes
		return fmt.Sprintf("MapOf%sTo%s", gen.refToId(elemTypes[0]), gen.refToId(elemTypes[1]))
	case types.RefKind:
		return fmt.Sprintf("RefOf%s", gen.refToId(rt.Desc.(types.CompoundDesc).ElemTypes[0]))
	case types.SetKind:
		return fmt.Sprintf("SetOf%s", gen.refToId(rt.Desc.(types.CompoundDesc).ElemTypes[0]))
	case types.StructKind:
		// We get an empty name when we have a struct that is used as union
		if rt.Name() == "" {
			choices := rt.Desc.(types.StructDesc).Union
			s := "__unionOf"
			for i, f := range choices {
				if i > 0 {
					s += "And"
				}
				s += strings.Title(f.Name) + "Of" + gen.refToId(f.T)
			}
			return s
		}

		if t.HasPackageRef() {
			return ToTag(t.PackageRef()) + "." + rt.Name()
		}
		return rt.Name()
	}
	panic("unreachable")
}

func (gen Generator) refToId(t types.TypeRef) string {
	if !t.IsUnresolved() || !t.HasPackageRef() {
		return gen.UserName(t)
	}
	rt := gen.R.Resolve(t)
	return fmt.Sprintf("%s_%s", ToTag(t.PackageRef()), gen.UserName(rt))
}

// ToTypeRef returns a string containing Go code that instantiates a types.TypeRef instance equivalent to t.
func (gen Generator) ToTypeRef(t types.TypeRef, fileID, packageName string) string {
	if t.HasPackageRef() {
		return fmt.Sprintf(`types.MakeTypeRef(ref.Parse("%s"), %d)`, t.PackageRef().String(), t.Ordinal())
	}
	if t.IsUnresolved() && fileID != "" {
		return fmt.Sprintf(`types.MakeTypeRef(__%sPackageInFile_%s_CachedRef, %d)`, packageName, fileID, t.Ordinal())
	}
	if t.IsUnresolved() {
		d.Chk.True(t.HasOrdinal(), "%s does not have an ordinal set", t.Name())
		return fmt.Sprintf(`types.MakeTypeRef(ref.Ref{}, %d)`, t.Ordinal())
	}

	if types.IsPrimitiveKind(t.Kind()) {
		return fmt.Sprintf("types.MakePrimitiveTypeRef(types.%sKind)", kindToString(t.Kind()))
	}
	switch desc := t.Desc.(type) {
	case types.CompoundDesc:
		typerefs := make([]string, len(desc.ElemTypes))
		for i, t := range desc.ElemTypes {
			typerefs[i] = gen.ToTypeRef(t, fileID, packageName)
		}
		return fmt.Sprintf(`types.MakeCompoundTypeRef("%s", types.%sKind, %s)`, t.Name(), kindToString(t.Kind()), strings.Join(typerefs, ", "))
	case types.EnumDesc:
		return fmt.Sprintf(`types.MakeEnumTypeRef("%s", "%s")`, t.Name(), strings.Join(desc.IDs, `", "`))
	case types.StructDesc:
		flatten := func(f []types.Field) string {
			out := make([]string, 0, len(f))
			for _, field := range f {
				out = append(out, fmt.Sprintf(`types.Field{"%s", %s, %t},`, field.Name, gen.ToTypeRef(field.T, fileID, packageName), field.Optional))
			}
			return strings.Join(out, "\n")
		}
		fields := fmt.Sprintf("[]types.Field{\n%s\n}", flatten(desc.Fields))
		choices := fmt.Sprintf("types.Choices{\n%s\n}", flatten(desc.Union))
		return fmt.Sprintf("types.MakeStructTypeRef(\"%s\",\n%s,\n%s,\n)", t.Name(), fields, choices)
	default:
		d.Chk.Fail("Unknown TypeDesc.", "%#v (%T)", desc, desc)
	}
	panic("Unreachable")
}

// ToTag replaces "-" characters in s with "_", so it can be used in a Go identifier.
// TODO: replace other illegal chars as well?
func ToTag(r ref.Ref) string {
	return strings.Replace(r.String(), "-", "_", -1)
}

func kindToString(k types.NomsKind) (out string) {
	out = types.KindToString[k]
	d.Chk.NotEmpty(out, "Unknown NomsKind %d", k)
	return
}
