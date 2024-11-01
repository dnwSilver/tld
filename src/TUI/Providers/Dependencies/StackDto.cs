using System.Runtime.Serialization;
using YamlDotNet.Serialization;

namespace TUI.Providers.Dependencies;

[DataContract]
[YamlSerializable]
public class StackDto
{
    [YamlMember] [DataMember] public string Name { get; set; }

    [YamlMember] [DataMember] public string Icon { get; set; }

    [YamlMember] [DataMember] public DependencyDto[] Conventions { get; set; }

    [YamlMember] [DataMember] public IEnumerable<HubDto> Hubs { get; set; }
}