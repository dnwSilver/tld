using System.Runtime.Serialization;
using YamlDotNet.Serialization;

namespace TUI.Providers.Dependencies;

[DataContract]
[YamlSerializable]
public class DependenciesDto
{
    [DataMember] [YamlMember] public StackDto[] Stacks { get; set; }
}