using UnityEngine;
using D = UnityEngine.Debug;
using Jsonif;

public class Test : MonoBehaviour
{
    T Identify<T>(T v)
    {
        var vs = Json.ToJson(v);
        var r = Json.FromJson<T>(vs);
        var rs = Json.ToJson(r);
        D.Assert(v.Equals(r));
        D.Assert(v.GetHashCode() == r.GetHashCode());
        D.Assert(vs == rs);
        return Json.FromJson<T>(Json.ToJson(v));
    }

    void TestEmpty()
    {
        var a = new Empty.Test();
        a = Identify(a);
    }

    void TestMessage()
    {
        var a = new Message.Person();
        D.Assert(a.name == "");
        D.Assert(a.flag == false);
        a = Identify(a);
        D.Assert(a.name == "");
        D.Assert(a.flag == false);

        a.name = "foo";
        a.flag = true;
        a = Identify(a);
        D.Assert(a.name == "foo");
        D.Assert(a.flag == true);
    }


    void TestEnumpb()
    {
        var a = new Enumpb.Data();
        D.Assert(a == Enumpb.Data.FOO);
        a = Identify(a);
        D.Assert(a == Enumpb.Data.FOO);

        a = Enumpb.Data.BAR;
        a = Identify(a);
        D.Assert(a == Enumpb.Data.BAR);
    }

    void TestNested()
    {
        var a = new Nested.Nested.Test2();
        D.Assert(a.nested_message.name == "");
        D.Assert(a.nested_enum == Nested.Nested.Test.NestedEnum.FOO);
        D.Assert(a.test.nested_message.name == "");
        D.Assert(a.test.nested_enum == Nested.Nested.Test.NestedEnum.FOO);
        a = Identify(a);
        D.Assert(a.nested_message.name == "");
        D.Assert(a.nested_enum == Nested.Nested.Test.NestedEnum.FOO);
        D.Assert(a.test.nested_message.name == "");
        D.Assert(a.test.nested_enum == Nested.Nested.Test.NestedEnum.FOO);

        a.nested_message.name = "foo";
        a.nested_enum = Nested.Nested.Test.NestedEnum.BAR;
        a.test.nested_message.name = "bar";
        a.test.nested_enum = Nested.Nested.Test.NestedEnum.HOGE;
        a = Identify(a);
        D.Assert(a.nested_message.name == "foo");
        D.Assert(a.nested_enum == Nested.Nested.Test.NestedEnum.BAR);
        D.Assert(a.test.nested_message.name == "bar");
        D.Assert(a.test.nested_enum == Nested.Nested.Test.NestedEnum.HOGE);
    }

    void TestRepeated()
    {
        var a = new Repeated.Test();
        D.Assert(a.a.Count == 0);
        D.Assert(a.b.Count == 0);
        D.Assert(a.c.Count == 0);
        D.Assert(a.d.Count == 0);
        a = Identify(a);
        D.Assert(a.a.Count == 0);
        D.Assert(a.b.Count == 0);
        D.Assert(a.c.Count == 0);
        D.Assert(a.d.Count == 0);

        a.a.Add(1);
        a.b.Add("foo");
        a.c.Add(Repeated.Enum.BAR);
        a.d.Add(new Repeated.Message() { name = "bar" });
        a = Identify(a);
        D.Assert(a.a.Count == 1 && a.a[0] == 1);
        D.Assert(a.b.Count == 1 && a.b[0] == "foo");
        D.Assert(a.c.Count == 1 && a.c[0] == Repeated.Enum.BAR);
        D.Assert(a.d.Count == 1 && a.d[0].name == "bar");
    }

    void TestOneof()
    {
        var a = new Oneof.Test();
        D.Assert(a.test_oneof_case == Oneof.Test.TestOneofCase.NOT_SET);
        a = Identify(a);
        D.Assert(a.test_oneof_case == Oneof.Test.TestOneofCase.NOT_SET);

        a.SetA(1);
        D.Assert(a.test_oneof_case == Oneof.Test.TestOneofCase.kA);
        D.Assert(a.a == 1);
        a = Identify(a);
        D.Assert(a.test_oneof_case == Oneof.Test.TestOneofCase.kA);
        D.Assert(a.a == 1);

        a.SetB("foo");
        D.Assert(a.test_oneof_case == Oneof.Test.TestOneofCase.kB);
        D.Assert(a.b == "foo");
        a = Identify(a);
        D.Assert(a.test_oneof_case == Oneof.Test.TestOneofCase.kB);
        D.Assert(a.b == "foo");

        a.SetC(Oneof.Enum.BAR);
        D.Assert(a.test_oneof_case == Oneof.Test.TestOneofCase.kC);
        D.Assert(a.c == Oneof.Enum.BAR);
        a = Identify(a);
        D.Assert(a.test_oneof_case == Oneof.Test.TestOneofCase.kC);
        D.Assert(a.c == Oneof.Enum.BAR);

        a.SetD(new Oneof.Message() { name = "bar" });
        D.Assert(a.test_oneof_case == Oneof.Test.TestOneofCase.kD);
        D.Assert(a.d.name == "bar");
        a = Identify(a);
        D.Assert(a.test_oneof_case == Oneof.Test.TestOneofCase.kD);
        D.Assert(a.d.name == "bar");

        a.ClearTestOneofCase();
        D.Assert(a.test_oneof_case == Oneof.Test.TestOneofCase.NOT_SET);
    }

    void TestOptional()
    {
        var a = new Optional.Test();
        D.Assert(!a.HasA());
        D.Assert(!a.HasB());
        D.Assert(!a.HasC());
        D.Assert(!a.HasD());
        a = Identify(a);
        D.Assert(!a.HasA());
        D.Assert(!a.HasB());
        D.Assert(!a.HasC());
        D.Assert(!a.HasD());

        a.SetA(1);
        D.Assert(a.HasA());
        D.Assert(a.a == 1);
        a = Identify(a);
        D.Assert(a.HasA());
        D.Assert(a.a == 1);

        a.SetB("foo");
        D.Assert(a.HasB());
        D.Assert(a.b == "foo");
        a = Identify(a);
        D.Assert(a.HasB());
        D.Assert(a.b == "foo");

        a.SetC(Optional.Enum.BAR);
        D.Assert(a.HasC());
        D.Assert(a.c == Optional.Enum.BAR);
        a = Identify(a);
        D.Assert(a.HasC());
        D.Assert(a.c == Optional.Enum.BAR);

        a.SetD(new Optional.Message() { name = "bar" });
        D.Assert(a.HasD());
        D.Assert(a.d.name == "bar");
        a = Identify(a);
        D.Assert(a.HasD());
        D.Assert(a.d.name == "bar");

        a.ClearA();
        D.Assert(!a.HasA());
        a.ClearB();
        D.Assert(!a.HasB());
        a.ClearC();
        D.Assert(!a.HasC());
        a.ClearD();
        D.Assert(!a.HasD());
    }

    void TestImporting()
    {
        var a = new Importing.Test();
        D.Assert(a.t.nanos == 0);
        a = Identify(a);
        D.Assert(a.t.nanos == 0);
    }

    void Start()
    {
        TestEmpty();
        TestMessage();
        TestEnumpb();
        TestNested();
        TestRepeated();
        TestOneof();
        TestImporting();

        Debug.Log("Unity Test passed");
    }
}
