import type { FunctionalComponent } from 'vue';
import { h } from 'vue';

interface IconProps {
  fillColor?: string;
  strokeColor?: string;
  strokeWidth?: number | string;
}

const OrderIcon: FunctionalComponent<IconProps> = (props) => {
  const {
    fillColor = 'transparent',
    strokeColor = 'currentColor',
    strokeWidth = 1.5,
  } = props;

  return h(
    'svg',
    {
      xmlns: 'http://www.w3.org/2000/svg',
      width: '24',
      height: '24',
      viewBox: '0 0 24 24',
      fill: 'none',
    },
    [
      h('path', {
        d: 'M9 3H5a2 2 0 0 0-2 2v4m6-6h10a2 2 0 0 1 2 2v4M9 3v18m0 0h10a2 2 0 0 0 2-2V9M9 21H5a2 2 0 0 1-2-2V9m0 0h18',
        stroke: strokeColor,
        'stroke-width': strokeWidth,
        'stroke-linecap': 'round',
        'stroke-linejoin': 'round',
        fill: fillColor,
      }),
    ],
  );
};

OrderIcon.displayName = 'OrderIcon';

export default OrderIcon;
